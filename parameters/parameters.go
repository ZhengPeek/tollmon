package parameters

import (
	"os"
	"sync"
	"time"
	"tollsys/tollmon/datastruct"
	"tollsys/tollmon/db"
	"tollsys/tollmon/g"
	"tollsys/tollmon/redis"
)

var (
	lock     *sync.Mutex
	result   *db.QueryResultNodeList
	stations []datastruct.Node
	plazas   []datastruct.Node
	lanes    []datastruct.Node

	StationNodes []datastruct.Station
	PlazaNodes   []datastruct.Plaza

	ipToNode       map[string]datastruct.Node
	laneInfo       map[string]datastruct.LaneInfo
	strategyItems  []datastruct.StrategyItem
	typeToStrategy map[int]datastruct.StrategyItem

	coreInfo map[string]datastruct.CoreData

	laneQueue map[string]time.Time
)

func GetStationTrees() []datastruct.Station {
	return StationNodes
}
func GetPlazaTrees() []datastruct.Node {
	return plazas
}
func init() {
	lock = &sync.Mutex{}
	result = &db.QueryResultNodeList{}
	stations = []datastruct.Node{}
	plazas = []datastruct.Node{}
	lanes = []datastruct.Node{}
	StationNodes = []datastruct.Station{}
	PlazaNodes = []datastruct.Plaza{}

	ipToNode = make(map[string]datastruct.Node)
	laneInfo = make(map[string]datastruct.LaneInfo)
	coreInfo = make(map[string]datastruct.CoreData)

	strategyItems = make([]datastruct.StrategyItem, 0)
	typeToStrategy = make(map[int]datastruct.StrategyItem)
	laneQueue = make(map[string]time.Time)
}

//初始化Parameters模块
func InitParameters() {
	g.LogInfo("Init Parameters...")
	loadNodeTree()
	loadNodeMap()
	loadLaneInfo()
	loadCoreInfo()
	loadStrategyItems()
	g.LogInfo("Init Parameters OK...")
}
func loadStations() {
	sql := "select nodeID,nodeName,nodeIp,tranMode from nodeCode where substring(nodeId,13,2)='" + g.Config().Node.ID[6:8] + "' and right(nodeId,1)='5'"
	err := db.Client.ExecQuery(sql, result)
	if err != nil {
		g.LogError("load stations err:", err.Error())
		os.Exit(2)
	}
	stations, err = result.GetNodeList()
	result.GC()
}
func loadPlazas() {
	sql := "select nodeID,nodeName,nodeIp,tranMode from nodeCode where substring(nodeId,13,2)='" + g.Config().Node.ID[6:8] + "' and right(nodeId,1)='6'"
	err := db.Client.ExecQuery(sql, result)
	if err != nil {
		g.LogError("load stations err:", err.Error())
		os.Exit(2)
	}
	plazas, _ = result.GetNodeList()
	result.GC()
}
func loadLanes() {
	sql := "select nodeID,nodeName,nodeIp,tranMode from nodeCode where substring(nodeId,13,2)='" + g.Config().Node.ID[6:8] + "' and right(nodeId,1)='7'"
	err := db.Client.ExecQuery(sql, result)
	if err != nil {
		g.LogError("load stations err:", err.Error())
		os.Exit(2)
	}
	lanes, _ = result.GetNodeList()
	result.GC()
}

//为NodeTree赋值，创建站-广场-车道层级关系
func loadNodeTree() {
	loadStations()
	loadPlazas()
	loadLanes()
	for _, station := range stations {
		stationNode := datastruct.Station{Station: station}
		for _, plaza := range plazas {
			plazaToStation := datastruct.Plaza{Plaza: plaza}
			if plaza.NodeID[12:16] == station.NodeID[12:16] {
				laneToPlaza := make([]datastruct.Node, 0)
				for _, lane := range lanes {
					if lane.NodeID[12:20] == plaza.NodeID[12:20] {
						laneToPlaza = append(laneToPlaza, lane)
					}
				}
				plazaToStation.Lanes = laneToPlaza
				stationNode.Plazas = append(stationNode.Plazas, plazaToStation)
				PlazaNodes = append(PlazaNodes, plazaToStation)
			}
		}
		StationNodes = append(StationNodes, stationNode)
	}
	g.LogInfo("load node info ok")
}
func loadNodeMap() {
	for _, station := range stations {
		ipToNode[station.NodeIP] = station
	}
	for _, plaza := range plazas {
		ipToNode[plaza.NodeIP] = plaza
	}
	for _, lane := range lanes {
		ipToNode[lane.NodeIP] = lane
	}
}
func loadCoreInfo() {
	for _, lane := range lanes {
		coreInfo[lane.NodeID] = datastruct.NewCoreData(lane)
	}
}
func loadLaneInfo() {
	for _, lane := range lanes {
		laneInfo[lane.NodeID] = datastruct.NewLaneInfo(lane)
	}
}
func loadStrategyItems() {
	b := redis.Get("Strategy")
	if len(b) == 0 {
		g.LogError("strategy items are blank")
		os.Exit(1)
	}
	err := g.Json.Unmarshal(b, &strategyItems)
	if err != nil {
		g.LogError("parse strategy items err :", err.Error())
		os.Exit(1)
	}

	for _, v := range strategyItems {
		typeToStrategy[v.Type] = v
	}
}
func GetNodeByIP(ip string) (*datastruct.Node, bool) {
	if laneNode, ok := ipToNode[ip]; ok {
		return &laneNode, true
	}
	return nil, false
}
func GetLaneInfoByIP(ip string) *datastruct.LaneInfo {
	if laneNode, ok := ipToNode[ip]; ok {
		info := laneInfo[laneNode.NodeID]
		return &info
	}
	return nil
}
func GetLaneInfoByID(id string) *datastruct.LaneInfo {
	lock.Lock()
	info := laneInfo[id]
	lock.Unlock()
	return &info
}
func UpdateLaneInfo(id string, key string, val interface{}) {
	lock.Lock()
	if _, ok := laneInfo[id]; ok {
		laneInfo[id].UpdateInfo(key, val)
	}
	lock.Unlock()
}
func UpdateCoreInfo(id string, key string, val interface{}) {
	lock.Lock()
	if _, ok := coreInfo[id]; ok {
		coreInfo[id].UpdateData(key, val)
	}
	lock.Unlock()
}
func GetCoreInfoById(id string) *datastruct.CoreData {
	data := coreInfo[id]
	return &data
}

func GetLaneInfoByStationID(id string) []datastruct.LaneInfo {
	list := make([]datastruct.LaneInfo, 0)
	for k, v := range laneInfo {
		if k[0:16] == id[0:16] {
			list = append(list, v)
		}
	}
	return list
}
func GetCoreDataByStationID(id string) []datastruct.CoreData {
	list := make([]datastruct.CoreData, 0)
	for k, v := range coreInfo {
		if k[0:16] == id[0:16] {
			list = append(list, v)
		}
	}
	return list
}

func GetStrategyItems() []datastruct.StrategyItem {
	return strategyItems
}
func GetTypeToStrategyItems() map[int]datastruct.StrategyItem {
	return typeToStrategy
}
func GetLaneQueue() map[string]time.Time {
	return laneQueue
}
