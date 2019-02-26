package g

import (
	"os"

	"runtime"

	log "github.com/cihub/seelog"
)

var (
	logger log.LogContextInterface
)

func init() {
	logger, err := log.LoggerFromConfigAsFile("./config/seelog.xml")
	if err != nil {
		log.Critical("err parsing config log file", err)
		os.Exit(0)
	}
	log.ReplaceLogger(logger)
}
func LogFlush() {
	log.Flush()
}

//LogInfo info日志
func LogInfo(v ...interface{}) {
	pc, _, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc).Name()
	log.Info("[", f, "]", " [", line, "] ", v)
}

//LogDebug debug日志
func LogDebug(v ...interface{}) {
	if !Config().Log.Debug {
		return
	}
	pc, _, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc).Name()
	log.Debug("[", f, "]", " [", line, "] ", v)
}

//LogError err日志
func LogError(v ...interface{}) {
	pc, _, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc).Name()
	log.Error("[", f, "]", " [", line, "] ", v)
}
