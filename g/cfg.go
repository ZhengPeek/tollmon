package g

import (
	"os"
	"sync/atomic"
	"unsafe"

	"fmt"

	"github.com/json-iterator/go"
	"github.com/toolkits/file"
)

type LogConfig struct {
	Debug bool `json:"debug"`
}
type NodeConfig struct {
	ID string `json:"id"`
	IP string `json:"ip"`
}
type HttpConfig struct {
	Listen string `json:"listen"`
}
type WebSocketConfig struct {
	Listen string `json:"listen"`
	Interval int `json:"interval"`
}
type MonitorConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}
type RedisConfig struct {
	ConnectType string `json:"connectType"`
	Host        string `json:"host"`
	MaxPoolSize int    `json:"maxPoolSize"`
}
type DBConfig struct {
	Host        string `json:"host"`
	User        string `json:"user"`
	DbName      string `json:"dbName"`
	Pwd         string `json:"pwd"`
	MaxPoolSize int    `json:"maxPoolSize"`
}
type SessionConfig struct {
	CookieName string `json:"cookieName"`
	MaxAge     int    `json:"maxAge"`
}
type RoadConfig struct {
	Node string `json:"node"`
}
type CoreDataConfig struct {
	List map[string]int `json:"list"`
}
type GlobalConfig struct {
	Log       *LogConfig       `json:"log"`
	Node      *NodeConfig      `json:"node"`
	Http      *HttpConfig      `json:"http"`
	WebSocket *WebSocketConfig `json:"webSocket"`
	DB        *DBConfig        `json:"db"`
	Redis     *RedisConfig     `json:"redis"`
	Monitor   *MonitorConfig   `json:"monitor"`
	Session   *SessionConfig   `json:"session"`
	CoreData  *CoreDataConfig  `json:"coredata"`
}

var (
	ptr  unsafe.Pointer
	Json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func Config() *GlobalConfig {
	return (*GlobalConfig)(atomic.LoadPointer(&ptr))
}

//config.json 序列化为GlobalConfig实例
func ParseConfig(cfg string) {
	if cfg == "" {
		LogError("config file not exists")
		os.Exit(1)
	}
	if !file.IsExist(cfg) {
		LogError("cfg file is not found ", cfg)
		os.Exit(1)
	}

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		LogError("read cfg err ", cfg, "errMsg:", err.Error())
		os.Exit(1)
	}

	var c GlobalConfig
	err = Json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		LogError("parse cfg err ", err.Error())
		fmt.Println(err.Error())
		os.Exit(1)
	}
	//保存实例指针至全局变量
	atomic.StorePointer(&ptr, unsafe.Pointer(&c))

	LogInfo("parese cfg file OK")
}
