package log

import (
	"github.com/expgo/config"
	"testing"
)

type MyLogStruct struct {
}

type MyLog struct {
}

func TestLog(t *testing.T) {
	log := Log[MyLogStruct]()
	log.Info("hello")
}

func TestLevel(t *testing.T) {
	cfg, _ := config.New[Config]("")
	cfg.Level["*Struct"] = LevelDebug
	cfg.File.Filename = "app.log"

	structLog, _ := NewWithConfig[MyLogStruct](cfg)
	log := Log[MyLog]()

	structLog.Debug("debug hello")
	log.Debug("debug hello")

	structLog.Info("info hello")
	log.Info("info hello")
}
