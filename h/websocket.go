package h

import (
	"net/http"
	"sync"
	"time"
	"tollsys/tollmon/datastruct"
	"tollsys/tollmon/g"
	"tollsys/tollmon/parameters"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type webSocketClient struct {
	lock          *sync.Mutex
	client        *websocket.Conn
	requestIds    map[string]bool
	strategyItems map[int]datastruct.StrategyItem
	stop          chan bool
}

func (w *webSocketClient) write(i interface{}) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.client.WriteJSON(i)
}
func newWebSocketClient() webSocketClient {
	return webSocketClient{lock: &sync.Mutex{}, requestIds: make(map[string]bool, 0),
		strategyItems: make(map[int]datastruct.StrategyItem),
		stop:          make(chan bool)}
}

var (
	upGrader   websocket.Upgrader
	clientList = make(map[*webSocketClient]bool)
)

//初始化webSocket升级方法，允许跨域访问
func init() {
	upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}
}

//配置webSocket路由
func configWebSocketHandle() {
	v1Ws.GET("/ws", func(context *gin.Context) {
		wsHandle(context)
	})
}

//webSocket出路模块
//获取webSocket连接并通过cookie获取session中请求收费站信息和报警策略
//绑定webSocket-请求站点；webSocket-报警策略
//通过goroutine启动实时数据处理模块
//阻塞操作-心跳检测发送停止信号时将终止该goroutine
func wsHandle(c *gin.Context) {
	g.LogDebug("handle webSocket...")
	conn := newWebSocketClient()
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		g.LogError("ws upgrade err:", err.Error())
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	defer ws.Close()
	g.LogDebug("webSocket conn from ", ws.RemoteAddr())
	session := Manager.GetSession(c)
	if session.Data == nil {
		lock.Lock()
		ws.WriteJSON("nil session")
		lock.Unlock()
		return
	}
	conn.client = ws
	temp := session.Data["requestIds"]
	bTemp, _ := g.Json.Marshal(temp)
	ids := make([]string, 0)
	err = g.Json.Unmarshal(bTemp, &ids)
	if err != nil {
		g.LogError(err)
		return
	}
	if len(ids) != 0 {
		for _, stationIds := range ids {
			conn.requestIds[stationIds] = true
		}
	} else {
		conn.client.WriteJSON("nil requestIds")
		conn.client.Close()
		return
	}
	g.LogInfo(conn.client.RemoteAddr(), ":requestIds - ", conn.requestIds)

	temp = session.Data["StrategyItems"]
	bTemp, _ = g.Json.Marshal(temp)
	//fmt.Println(string(bTemp))
	items := make(map[int]datastruct.StrategyItem)
	err = g.Json.Unmarshal(bTemp, &items)
	if err != nil {
		g.LogError(err)
		return
	}
	if len(items) != 0 {
		conn.strategyItems = items
	} else {
		conn.strategyItems = parameters.GetTypeToStrategyItems()
	}
	g.LogInfo(conn.client.RemoteAddr(), ":strategyItems - ", conn.strategyItems)

	clientList[&conn] = true
	exit := false
	go func() {
		if exit {
			return
		}
		a := datastruct.NewCommonMessage()
		err := conn.client.ReadJSON(a)
		if err != nil {
			g.LogDebug("ws read err:", err.Error())
		}
		if a.Data == "close" {
			clientList[&conn] = false
		}
	}()
	select {
	case <-conn.stop:
		lock.Lock()
		delete(clientList, &conn)
		exit = true
		g.LogInfo(conn.client.RemoteAddr(), " get stop signal")
		lock.Unlock()
		g.LogDebug(conn.client.RemoteAddr(), "已终止")
		return
	}
}

//webSocketHeartServ webSocket客户端心跳检测
//根据配置定期发送0，保证长连接有效
//遍历webSocket客户端列表，状态为true则发送心跳检测 - 成功：继续轮询 ; 失败：置状态为false并关闭该goroutine
//若状态为false则该webSocket已失效，直接关闭该goroutine
func webSocketHeartServ() {
	g.LogInfo("goroutine start - webSocket heart beat")
	for {
		for conn, ok := range clientList {
			if ok {
				err := conn.write(0)
				if err != nil {
					g.LogDebug("heart beat err:", err.Error())
					conn.client.Close()
					conn.stop <- true
					g.LogDebug("webSocket", conn.client.RemoteAddr(), "停止信号已发送")
				}
				g.LogDebug("webSocket心跳检测-", conn.client.RemoteAddr(), "成功")
			} else {
				conn.stop <- true
			}
		}
		time.Sleep(time.Second * time.Duration(g.Config().WebSocket.Interval))
	}
}
