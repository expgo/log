package example

import "github.com/expgo/log"

// @Log
//go:generate ag --dev-plugin=github.com/expgo/log/annotation

type MyLog struct {
	log.InnerLog
}

func (ml *MyLog) WriteLog() {
	ml.L.Debug("debug")
	ml.L.Info("info")
	ml.L.Warn("warn")
	ml.L.Error("error")
}

type MyLog1 struct {
	log log.Logger `new:"self,value:log1"`
}

func (ml *MyLog1) WriteLog() {
	ml.log.Debug("debug 1")
	ml.log.Info("info 2")
	ml.log.Warn("warn 3")
	ml.log.Error("error 4")
}
