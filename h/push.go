package h

import (
	"net/http"
	"time"
	"tollsys/tollmon/datastruct"
	"tollsys/tollmon/g"
	"tollsys/tollmon/parameters"

	"github.com/gin-gonic/gin"
)

//ConfigPushHandle 路由配置
func configPushHandle() {
	//push POST 接收发布端的POST信息更新渲染数据并添加至实时发送队列
	v1.POST("/push", func(c *gin.Context) {
		if c.Request.ContentLength == 0 {
			c.JSON(http.StatusBadRequest, datastruct.ERRORMSG_BlankBody)
			c.Abort()
		}
		decoder := g.Json.NewDecoder(c.Request.Body)
		var metrics []datastruct.MetricValue
		err := decoder.Decode(&metrics)
		if err != nil {
			g.LogDebug(err.Error())
			c.JSON(http.StatusBadRequest, datastruct.ERRORMSG_DecoderError)
			c.Abort()
			return
		}
		for _, m := range metrics {
			msgSend := datastruct.NewMsgSend()
			node, exists := parameters.GetNodeByIP(m.Endpoint)
			if !exists {
				g.LogError(m.Endpoint, " is not exists")
				continue
			}
			if msgType, ok := g.Config().CoreData.List[m.Metric]; ok {
				parameters.UpdateCoreInfo(node.NodeID, m.Metric, m.Value)
				msgSend.MsgCatalog = 22
				msgSend.MsgType = msgType
				msgSend.MsgTime = time.Unix(m.Timestamp, 0).Format("2006-01-02 15:04:05")
				msgSend.MsgLane = node.NodeID
				a := make(map[string]interface{})
				a[m.Metric] = m.Value
				msgSend.MsgContent = a
				PushRealData(node.NodeID[0:16], msgSend)
			}
		}

	})
}
