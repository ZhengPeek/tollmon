package t

import (
	"tollsys/tollmon/h"
	"time"
	"tollsys/tollmon/g"
	"tollsys/tollmon/db"
	"tollsys/tollmon/datastruct"
	"github.com/toolkits/file"
	"os"
	"tollsys/tollmon/redis"
)

func HandleMessage() {
	i := 0
	for {
		msg := datastruct.NewCommonMessage()
		msg.Data = time.Now()
		i++
		h.PushRealData("1F01000000000401",msg)
		g.LogDebug("insert test data ok")
		time.Sleep(time.Duration(2) * time.Second)
	}
}
func TestStrategyItems(){
	configContent, err := file.ToTrimString("./config/strategyitems.json")
	if err != nil{
		g.LogDebug(err.Error())
		os.Exit(1)
	}
	var items []datastruct.StrategyItem
	err = g.Json.Unmarshal([]byte(configContent),&items)
	if err != nil{
		g.LogDebug(err.Error())
		os.Exit(1)
	}
	g.LogDebug(items)
	val,_ := g.Json.Marshal(items)
	redis.Set("Strategy",val)
	b := redis.Get("Strategy")
	items = nil
	err = g.Json.Unmarshal(b,&items)
	if err != nil{
		g.LogDebug(err.Error())
		os.Exit(1)
	}
	g.LogDebug(items)
}
type TestJson struct {
	Testfield   int    `json:"testfield"`
	Stringfield string `json:"stringfield"`
}

func TestSessionStorage() {
	for {
		h.Manager.TestSession()
		time.Sleep(time.Duration(2 * time.Second))
	}
}
func TestDB() {
	db.InitDB()
	var result = &db.QueryResultNodeList{

	}
	db.Client.ExecQuery("select top 5 nodeid,nodename,nodeip from nodecode where right(nodeid,1) = '7'", result)
	g.LogInfo(result.GetNodeList())
}
