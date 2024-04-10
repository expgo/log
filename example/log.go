package example

import "github.com/expgo/log"

//go:generate ag --dev-plugin=github.com/expgo/log/annotation --dev-plugin-dir=../

// @Log
type MyLog struct {
	log *log.Logger[MyLog] `wire:"auto"`
}

func (ml *MyLog) WriteLog() {
	ml.log.Debug("debug")
	ml.log.Info("info")
	ml.log.Warn("warn")
	ml.log.Error("error")
}
