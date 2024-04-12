package example

import "github.com/expgo/log"

// @Log
//go:generate ag --dev-plugin=github.com/expgo/log/annotation

type MyLog struct {
	log log.Logger `new:""`
}

func (ml *MyLog) WriteLog() {
	ml.log.Debug("debug")
	ml.log.Info("info")
	ml.log.Warn("warn")
	ml.log.Error("error")
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
