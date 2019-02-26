package h

import (
	"time"
	"tollsys/tollmon/datastruct"
	"tollsys/tollmon/g"

	"github.com/gorilla/websocket"
	"runtime"
)

var (
	broadCast        = make(map[string]chan interface{})
	stationToClients = make(map[string][]*websocket.Conn)
)

//bindStationIdToClients 提供绑定 站节点-webSocket方法
//已废弃
//func bindStationIdToClients(id string, conn *websocket.Conn) {
//	stationToClients[id] = append(stationToClients[id], conn)
//}

//bindStationIdToClients 提供解除绑定 站节点-webSocket方法
//遍历队列获取失效webSocket索引，合并该索引前后的队列为新的队列
//func unbindStationIdToClient(id string, conn *websocket.Conn) {
//	for index, client := range stationToClients[id] {
//		if client == conn {
//			g.LogDebug("解除绑定 站节点-webSocket", conn.RemoteAddr(), "-", id)
//			stationToClients[id] = append(stationToClients[id][:index], stationToClients[id][index+1:]...)
//		}
//	}
//}
func unbindStationIdToClient(conn *websocket.Conn) {
	for index, webSocketList := range stationToClients {
		for id, client := range webSocketList {
			if client == conn {
				g.LogDebug("解除绑定 站节点-webSocket", conn.RemoteAddr(), "-", id)
				stationToClients[index] = append(stationToClients[index][:id], stationToClients[index][id+1:]...)
			}
		}
	}
}

//getDataChan提供通过stationID获取缓冲数据通道的功能
//要求参数stationID:string eg:(1F010000000008)
//如果缓冲列表中无该站缓冲则创建
func getDataChan(stationID string) chan interface{} {
	if dataChan, ok := broadCast[stationID]; ok {
		g.LogDebug("get chan:", stationID)
		return dataChan
	}
	g.LogDebug("chan:", stationID, " is blank , create a new chan")
	dataChan := make(chan interface{}, 5000)
	broadCast[stationID] = dataChan
	return dataChan
}
func PushRealData(stationId string,data interface{}){
	for conn,_ := range clientList { //遍历webSocket客户端列表
		if conn.requestIds[stationId] { //判断该webSocket客户端是否请求当前站数据并发送
			send(conn, data)
		}
	}
}
//PushRuntimeData提供写入实时数据至某收费站的方法
//要求参数stationID:string eg:(1F01000000008) 数据data:interface{}
//通过stationID获取该station的缓冲数据通道并将data写入缓冲通道
//当通道满时，去除最早的数据并写入最新数据
//func PushRuntimeData(stationID string, data interface{}) {
//	lock.Lock()
//	defer lock.Unlock()
//	c := getDataChan(stationID)
//	select {
//	case c <- data:
//		g.LogDebug("<- chan OK:", stationID, data)
//	default:
//		g.LogDebug("chan full")
//		d := <-c
//		g.LogDebug("remove index 1 data:", d)
//		d = nil
//		c <- data
//	}
//}

//StartRealData 广播实时数据 goroutine启动
//遍历待发送实时数据队列，获取站实时数据缓冲通道并根据站节点编码获取该通道需要该通道数据的webSocket 广播该通道数据
func StartRealData() {
	g.LogInfo("goroutine:Start Handle Real Data Send...")
	g.LogDebug("当前广播通道:",len(broadCast))
	for {
		for stationId, dataChan := range broadCast {
			g.LogDebug("开始广播信息:",stationId," - ", len(dataChan))
			//g.LogDebug("start broadcasting data-",stationId)
			//需要注意的是 待广播数据模型是 站-数据缓存队列 的模式，因此在一次发送中将发送缓存队列中所有数据
			//缓存数据的发送暂时无法做到并发发送，webSocket是不支持并发写操作的
			//因此在数据量剧增时，前端可能会出现短时间内仅接收到同一站点的信息，这是由于缓存数据过多导致的
			//如果需要做到前端同时获取多个站数据的效果，可以并发读取站数据再写入webSocket，由于写操作由读写锁控制
			//效率上并无提升，但用户体验将变好
			select {
				case data := <-dataChan: //从缓冲队列拿一条数据
					for conn, _ := range clientList { //遍历webSocket客户端列表
						if conn.requestIds[stationId] { //判断该webSocket客户端是否请求当前站数据并发送
							send(conn, data)
						}
					}
					data =  nil
				default: //缓存为空终止IO阻塞
					break
				}
		}
		time.Sleep(time.Second / 10)
		runtime.GC()
	}
}

//发送数据，并为该数据叠加策略信息
//发送失败则认为该webSocket已失效 (失效后仅标记该webSocket为失效状态，不做删除操作)
func send(conn *webSocketClient, d interface{}) {
	if conn.client == nil {
		return
	}
	if v, ok := d.(datastruct.MsgSend); ok {
		if v.MsgCatalog == 0x20 {
			if !conn.strategyItems[v.MsgType].IsChecked {
				return
			}
			v.MsgContent["level"] = conn.strategyItems[v.MsgType].Level
		}
	}
	j := datastruct.NewCommonMessage()
	j.Data = d
	if err := conn.write(j); err != nil {
		lock.Lock()
		g.LogError("ws:", conn.client.RemoteAddr(), " write data err:", err.Error())
		clientList[conn] = false
		conn.client.Close()
		lock.Unlock()
	}
	g.LogDebug("发送数据 - ",d," --> ",conn.client.RemoteAddr())
}
