package datastruct

import "sync"

const(
	KEY_RequestIds = "requestIds"
	KEY_StrategyItems ="StrategyItems"
	ERRORMSG_DecoderError = "cannot decoder body"
	ERRORMSG_BlankBody = "body is blank"
)
//Node 节点信息
type Node struct {
	NodeID   string `json:"nodeID"`
	NodeName string `json:"nodeName"`
	NodeIP   string `json:"nodeIP"`
	NodeType int    `json:"nodeType"`
	TranMode int    `json:"tranMode"`
}

//Plaza 广场信息 包含该广场下所有车道
type Plaza struct {
	Plaza Node   `json:"plaza"`
	Lanes []Node `json:"lanes"`
}

//Station 收费站信息 包含该收费站下所有收费广场
type Station struct {
	Station Node    `json:"station"`
	Plazas  []Plaza `json:"plazas"`
}

//CommonMessage 与前端通讯公共结构
//Code标识自定义错误码 默认0
//ErrMsg标识错误信息
//Data标识本次通讯数据
//Status标识本次通讯状态
type CommonMessage struct {
	Code   int         `json:"code"`
	ErrMsg string      `json:"errMsg"`
	Data   interface{} `json:"data"`
	Status bool        `json:"status"`
}

//NewCommonMessage获取一个CommonMessage实例
func NewCommonMessage() *CommonMessage {
	return &CommonMessage{
		Code:   0,
		ErrMsg: "",
		Status: true,
	}
}

//前端交互数据结构
//MsgCatalog 消息种类
//MsgType 消息类别
//MsgTime 消息产生时间
//MsgLane 消息产生车道节点
//MsgContent 消息内容
type MsgSend struct {
	MsgCatalog int
	MsgType    int
	MsgTime    string
	MsgLane    string
	MsgContent map[string]interface{}
}

func NewMsgSend() MsgSend {
	return MsgSend{MsgContent: make(map[string]interface{}),}
}

//监控时间数据结构
type MetricValue struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

//车道信息数据结构
type LaneInfo struct {
	Node Node                   `json:"node"`
	Info map[string]interface{} `json:"info"`
	lock *sync.Mutex
}

//获取车道数据实例
func NewLaneInfo(node Node) LaneInfo {
	a := make(map[string]interface{})
	a["ConnectStatus"] = false
	return LaneInfo{Info: a, Node: node, lock: &sync.Mutex{}}
}

//更新车道实例数据
//参数：key string 更新项主键，val interface 更新项值
//并发场景下未避免读写不一致，该操作为同步锁操作
func (l LaneInfo) UpdateInfo(key string, val interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Info[key] = val
}

//获取车道实例下信息内容
func (l LaneInfo) GetInfo() map[string]interface{} {
	return l.Info
}

//车道核心数据结构
type CoreData struct {
	Node     Node                   `json:"node"`
	lock     *sync.Mutex
	CoreData map[string]interface{} `json:"coreData"`
}

//获取车道核心数据实例
func NewCoreData(node Node) CoreData {
	return CoreData{CoreData: make(map[string]interface{}), Node: node, lock: &sync.Mutex{}}
}

//更新核心数据项
//参数：key string 更新项主键，val interface 更新项值
//并发场景下未避免读写不一致，该操作为同步锁操作
func (c CoreData) UpdateData(key string, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.CoreData[key] = val
}

//获取核心数据实例
func (c CoreData) Data() map[string]interface{} {
	return c.CoreData
}

//报警策略数据结构
type StrategyItem struct {
	Type        int    `json:"type"`
	Description string `json:"description"`
	IsChecked   bool   `json:"isChecked"`
	Level       int    `json:"level"`
}
