package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"tollsys/tollmon/g"
	"tollsys/tollmon/redis"
	_ "github.com/cihub/seelog"

	_ "net/http/pprof"
	"tollsys/tollmon/db"
	"tollsys/tollmon/h"
	"tollsys/tollmon/monitor"
	"tollsys/tollmon/parameters"
	"net/http"
)

var (
	VERSION = "unknown"
	BUILD   = "unknown"

	showVer     bool
	setStrategy bool
)

func InitSys() {
	db.InitDB()
	redis.InitRedis()
	h.InitServer()
	monitor.InitMonitor()
	parameters.InitParameters()
}
func main() {
	flag.BoolVar(&showVer, "v", false, "")
	flag.BoolVar(&setStrategy, "s", false, "set strategy to redis")
	flag.Parse()
	if showVer {
		fmt.Println(VERSION)
		fmt.Println(BUILD)
		os.Exit(0)
	}
	defer func() {
		if p := recover(); p != nil {
			g.LogError("panic recover:", p)
			str, ok := p.(string)
			if ok {
				fmt.Println(str)
			}
			debug.PrintStack()
			g.LogFlush()
		}
	}()
	runtime.GOMAXPROCS(runtime.NumCPU())

	//g.LogInfo("test start...")
	g.ParseConfig("./config/config.json")
	InitSys()
	//t.TestStrategyItems()

	//t.TestStrategyItems()
	//t.TestDB()
	//t.TestSessionStorage()
	//g.LogDebug(g.Config().Http.Listen)
	if g.Config().Log.Debug {
		go func() {
			g.LogInfo("start pprof")
			http.ListenAndServe("127.0.0.1:16666", nil)
		}()
	}

	go monitor.Start()
	//go t.HandleMessage()
	h.Start()
	select {}
}
