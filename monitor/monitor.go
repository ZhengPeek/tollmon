package monitor

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"tollsys/tollmon/g"
	"tollsys/tollmon/h"
	"tollsys/tollmon/parameters"
)

var (
	MONITORADDR string
	lock        *sync.Mutex
)

func InitMonitor() {
	MONITORADDR = g.Config().Monitor.Host + ":" + strconv.Itoa(g.Config().Monitor.Port)
	lock = &sync.Mutex{}
}

func Start() {
	//Start 启动实时监控TCP服务端
	//监听config.json中配置的addr
	//go程启动，监听客户端连接转至客户端操作，此goroutine常驻
	go func() {
		g.LogInfo("Start Monitor Server:", MONITORADDR)
		netListen, err := net.Listen("tcp", MONITORADDR)
		if err != nil {
			g.LogError("Listening addr ", MONITORADDR, " ", err.Error())
		}
		defer netListen.Close()
		g.LogInfo("Waiting for clients")
		for {
			conn, err := netListen.Accept()
			if err != nil {
				continue
			}
			g.LogInfo(conn.RemoteAddr().String(), "tcp conn success")
			go handleConnection(conn)
		}
	}()
	//goroutine 车道队列连接状态轮询
	//遍历map获取车道id与最后一次通讯时间
	//车道状态变更时才更新渲染数据并添加实时变更数据至实时发送队列，此goroutine常驻
	go func() {
		g.LogInfo("goroutine start - [lane queue connected status check]")
		for {
			lock.Lock()
			for nodeId, lastCommTime := range parameters.GetLaneQueue() {
				//fmt.Println(parameters.GetLaneInfoByID(nodeId).Node.NodeName, time.Now().Format("2006-01-02 15:04:05"), lastCommTime.Format("2006-01-02 15:04:05"))
				if time.Now().Sub(lastCommTime) > 20*time.Second {
					if parameters.GetLaneInfoByID(nodeId).Info["ConnectStatus"].(bool) {
						parameters.UpdateLaneInfo(nodeId, "ConnectStatus", false)
						a := make(map[string]interface{})
						a["ConnectStatus"] = false
						msg := setMsgSend(McTest, MtHeart, lastCommTime.Format("2006-01-02 15:04:05"), nodeId, a)
						h.PushRealData(nodeId[:16], msg)
						g.LogInfo("车道连接状态变更:", parameters.GetLaneInfoByID(nodeId).Node.NodeName, " - 已中断连接")
					}
				} else {
					if !parameters.GetLaneInfoByID(nodeId).Info["ConnectStatus"].(bool) {
						parameters.UpdateLaneInfo(nodeId, "ConnectStatus", true)
						a := make(map[string]interface{})
						a["ConnectStatus"] = true
						msg := setMsgSend(McTest, MtHeart, lastCommTime.Format("2006-01-02 15:04:05"), nodeId, a)
						h.PushRealData(nodeId[:16], msg)
						g.LogInfo("车道连接状态变更:", parameters.GetLaneInfoByID(nodeId).Node.NodeName, " - 通讯连接已建立")
					}
				}
			}
			lock.Unlock()
			time.Sleep(time.Second / 10)
		}
	}()
}

//handleConnection 客户端连接处理方法
//参数要求：客户端连接实例
//接收报文并逐字节写入数据缓冲区，go程启动报文处理业务
//go程启动，当连接中断时终止go程;当接收到不被允许的接连是终止go程

//TODO 二期工作将构建车道队列管理，仿照WebSocket客户端管理模式，将实时监控中对车道的socket管理变更为基于车道socket-车道逻辑节点的队列管理模式
func handleConnection(conn net.Conn) {
	g.LogInfo("handle client conn:", conn.RemoteAddr())
	//TODO 发布前需启用IP过滤，非本站点持有车道的连接将被拒绝
	if !g.Config().Log.Debug {
		remoteHost := conn.RemoteAddr()
		ip := strings.Split(remoteHost.String(), ":")[0]
		_, exists := parameters.GetNodeByIP(ip)
		if !exists {
			g.LogDebug("receive an unallowed conn ", remoteHost, " access denied")
			conn.Close()
			return
		}
	}

	buffer := make([]byte, 2048)
	var chanRev = make(chan byte, 4096) //数据缓冲区 接收字节加入这个队列
	var chanDisConnect = make(chan int) //goroutine同步器 中断时写入数据终止该goroutine下其余的goroutine
	//TODO 后续需要更改为车辆队列管理，协议编解码需要做channel同步
	go parseMsg(chanRev, chanDisConnect)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			g.LogError(conn.RemoteAddr().String(), " read error:", err.Error())
			chanDisConnect <- 1
			close(chanDisConnect)
			return
		}
		i := 0
		for {
			chanRev <- buffer[i]
			i++
			if i == n {
				break
			}
		}
	}
}
