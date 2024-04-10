package example

import "github.com/expgo/log"

//go:generate ag --dev-plugin=github.com/expgo/log/annotation --dev-plugin-dir=../

// @Log
type MyLog struct {
	log log.Logger[MyLog] `wire:"auto"`
}

func (ml *MyLog) WriteLog() {
	ml.log.Debug("debug")
	ml.log.Info("info")
	ml.log.Warn("warn")
	ml.log.Error("error")
}

// @Log(cfgPath="log1")
type MyLog1 struct {
	log log.Logger[MyLog1] `wire:"auto"`
}

func (ml *MyLog1) WriteLog() {
	ml.log.Debug("debug 1")
	ml.log.Info("info 2")
	ml.log.Warn("warn 3")
	ml.log.Error("error 4")
}
