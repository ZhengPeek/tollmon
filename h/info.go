package h

import (
	"net/http"
	"tollsys/tollmon/datastruct"
	"tollsys/tollmon/g"
	"tollsys/tollmon/parameters"

	"github.com/gin-gonic/gin"
)

//configNodeInfoRoute 配置/v1/NodeInfo路由，GET访问权限
//返回前端Station信息
func configNodeInfoRoute() {
	v1.GET("/NodeInfo", sessionMiddleWare(), func(c *gin.Context) {
		n := parameters.GetStationTrees()
		sender := datastruct.NewCommonMessage()
		sender.Data = n
		c.JSON(http.StatusOK, sender)
	})
}

//configLaneInfoRoute LaneInfo项路由配置
func configLaneInfoRoute() {
	// LaneInfo路由 获取客户端已订阅的laneInfo 若无则返回null
	v1.GET("/LaneInfo", requestNilMiddleWare(), sessionMiddleWare(), func(c *gin.Context) {
		sender := datastruct.NewCommonMessage()
		reps := make([]datastruct.LaneInfo, 0)
		s := Manager.GetSession(c)
		var val interface{}
		val = s.Get(datastruct.KEY_RequestIds)
		switch t := val.(type) {
		case []interface{}:
			for _, id := range t {
				laneInfoList := parameters.GetLaneInfoByStationID(id.(string))
				for _, v := range laneInfoList {
					reps = append(reps, v)
				}
			}
		}
		//forDebug
		//TODO 发布前注释debug输出
		//val := parameters.GetLaneInfosByStationID("1F010000000004010000000005")
		//for _, v := range val {
		//	reps = append(reps, v)
		//}
		sender.Data = reps
		c.JSON(http.StatusOK, sender)
	})
}

//configStrategyItemsRoute 报警策略路由配置
func configStrategyItemsRoute() {
	//StrategyInfo GET 根据前端请求Cookie获取该前端上次请求的策列配置
	v1.GET("/StrategyInfo", requestNilMiddleWare(), sessionMiddleWare(), func(c *gin.Context) {
		sender := datastruct.NewCommonMessage()
		val := Manager.GetSession(c).Data[datastruct.KEY_StrategyItems]
		items := make([]interface{}, 0)
		switch t := val.(type) {
		case map[string]interface{}:
			for _, item := range t {
				items = append(items, item)
			}
		}
		if items == nil || len(items) == 0 {
			for _, item := range parameters.GetStrategyItems() {
				items = append(items, item)
			}
		}
		//TODO 发布前应确认deBug信息被注释
		//items := parameters.GetStrategyItems()

		//if len(items) == 0 {
		//	sender.ErrMsg = "strategy items are empty"
		//	sender.Code = 1
		//}
		sender.Data = items
		c.JSON(http.StatusOK, sender)
	})
	//StrategyInfo GET 根据前端请求Cookie更新该请求的报警策略
	v1.POST("/StrategyInfo", sessionMiddleWare(), func(c *gin.Context) {
		if c.Request.ContentLength == 0 {
			c.JSON(http.StatusBadRequest, datastruct.ERRORMSG_BlankBody)
			c.Abort()
			return
		}
		type Rcvd struct {
			Code   int                       `json:"code"`
			ErrMsg string                    `json:"errMsg"`
			Data   []datastruct.StrategyItem `json:"data"`
			Status bool                      `json:"status"`
		}
		decoder := g.Json.NewDecoder(c.Request.Body)
		r := Rcvd{}
		r.Data = make([]datastruct.StrategyItem, 0)
		err := decoder.Decode(&r)
		if err != nil {
			c.JSON(http.StatusBadRequest, datastruct.ERRORMSG_DecoderError)
			c.Abort()
			return
		}
		mapItems := make(map[int]datastruct.StrategyItem)
		for _, item := range r.Data {
			mapItems[item.Type] = item
		}
		Manager.Update(c, datastruct.KEY_StrategyItems, mapItems)
		g.LogInfo(c.Request.RemoteAddr, " already update strategy items", mapItems)
		sender := datastruct.NewCommonMessage()
		sender.Data = true
		c.JSON(http.StatusOK, sender)
	})
}

//configNodeChoseRoute NodeChose路由配置
func configNodeChoseRoute() {
	//NodeChose GET 根据前端请求的cookie获得对应请求上次配置的请求节点
	v1.GET("/NodeChose", requestNilMiddleWare(), sessionMiddleWare(), func(c *gin.Context) {
		sender := datastruct.NewCommonMessage()
		val := Manager.GetSession(c).Get(datastruct.KEY_RequestIds)
		switch t := val.(type) {
		case []interface{}:
			ids := t
			sender.Data = ids
		}
		c.JSON(http.StatusOK, sender)
	})
	//NodeChose GET 根据前端请求的cookie更新配置的请求节点
	v1.POST("/NodeChose", sessionMiddleWare(), func(c *gin.Context) {
		if c.Request.ContentLength == 0 {
			c.JSON(http.StatusBadRequest, datastruct.ERRORMSG_BlankBody)
			c.Abort()
			return
		}
		var rcvd datastruct.CommonMessage

		decoder := g.Json.NewDecoder(c.Request.Body)
		err := decoder.Decode(&rcvd)
		if err != nil {
			c.JSON(http.StatusBadRequest, datastruct.ERRORMSG_DecoderError)
			c.Abort()
		}
		paraIds := make([]string, 0)
		for _, data := range rcvd.Data.([]interface{}) {
			if id, ok := data.(string); ok {
				paraIds = append(paraIds, id[:16])
			}
		}
		Manager.Update(c, datastruct.KEY_RequestIds, paraIds)

		g.LogInfo(c.Request.RemoteAddr, " request stations ", paraIds)
		sender := datastruct.NewCommonMessage()
		sender.Data = true
		c.JSON(http.StatusOK, sender)
	})
}

//configCoreDataRoute CoreData路由配置
func configCoreDataRoute() {
	//CoreData GET 根据前端请求的Cookie获得该节点请求站点的coreData信息
	v1.GET("/CoreData", requestNilMiddleWare(), sessionMiddleWare(), func(c *gin.Context) {
		s := Manager.GetSession(c)
		sender := datastruct.NewCommonMessage()
		val := s.Get(datastruct.KEY_RequestIds)
		resp := make([]datastruct.CoreData, 0)
		switch t := val.(type) {
		case []interface{}:
			for _, id := range t {
				coreDataList := parameters.GetCoreDataByStationID(id.(string))
				for _, coreData := range coreDataList {
					resp = append(resp, coreData)
				}
			}
		}
		sender.Data = resp
		c.JSON(http.StatusOK, sender)
	})
}
