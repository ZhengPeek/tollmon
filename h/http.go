package h

import (
	"net/http"
	"os"
	"time"
	"tollsys/tollmon/g"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	router   *gin.Engine
	wsRouter *gin.Engine
	Manager  *SessionManager
	v1       *gin.RouterGroup
	v1Ws     *gin.RouterGroup
)

//初始化HTTP/WebSocket发服务器信息及session管理器
func InitServer() {
	Manager = NewSessionManager()
	configRoutes()
}

//配置初始化路由控制及路由
func configRoutes() {
	//TODO 发布前启用Gin Release模式
	if !g.Config().Log.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	wsRouter = gin.New() //WebSocket Engine
	router = gin.New()   //Http Engine
	v1 = router.Group("/v1")
	v1Ws = wsRouter.Group("/v1")

	v1.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           720 * time.Hour}))
	v1Ws.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           720 * time.Hour}), sessionMiddleWare())

	v1.Static("/tollmon", "./statics/")
	configWebSocketHandle()
	configNodeInfoRoute()
	configNodeChoseRoute()
	configLaneInfoRoute()
	configStrategyItemsRoute()
	configPushHandle()
	configCoreDataRoute()
}

//以goroutine启动http和webSocket服务器
func Start() {
	go func() {
		g.LogInfo("HTTP Server Run At ", g.Config().Http.Listen)
		err := router.Run(g.Config().Http.Listen)
		if err != nil {
			g.LogError("HTTP Run Error:", err.Error())
			os.Exit(1)
		}
	}()

	go func() {
		g.LogInfo("WebSocket Server Run At ", g.Config().WebSocket.Listen)
		err := wsRouter.Run(g.Config().WebSocket.Listen)
		if err != nil {
			g.LogError("WebSocket Run Error:", err.Error())
			os.Exit(1)
		}
	}()
	go webSocketHeartServ()
	//go StartRealData()
}

//中间件
//sessionMiddleWare  启用session管理
func sessionMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		Manager.BeginSession(c)
		//c.Next()
	}
}

//中间件
//requestNilMiddleWare session管理 无session则拒绝访问
func requestNilMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := Manager.GetSession(c)
		if s != nil {
			c.Next()
			return
		}
		c.JSON(http.StatusBadRequest, "null session")
		c.Abort()
		return
	}
}
