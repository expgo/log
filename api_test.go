package log

import (
	"github.com/expgo/config"
	"github.com/expgo/factory"
	"testing"
)

type MyLogStruct struct {
}

type MyLog struct {
	log Logger `new:""`
}

func (m *MyLog) RunLog() {
	m.log.Debug("debug")
	m.log.Info("info")
	m.log.Warn("warn")
	m.log.Error("error")
}

func TestLog(t *testing.T) {
	log := Log[MyLogStruct]()
	log.Info("hello")
}

func TestLevel(t *testing.T) {
	cfg, _ := config.New[Config]("")
	cfg.Level["*Struct"] = LevelDebug
	cfg.File.Filename = "log/app.log"

	structLog := LogWithConfig[MyLogStruct](cfg)
	log := Log[MyLog]()

	structLog.Debug("debug hello")
	log.Debug("debug hello")

	structLog.Info("info hello")
	log.Info("info hello")
}

func TestLogRoll(t *testing.T) {
	cfg, _ := config.New[Config]("")
	cfg.Level["*Struct"] = LevelDebug
	cfg.File.Filename = "log/roll.log"
	cfg.File.MaxSize = 1
	cfg.File.MaxBackups = 3
	cfg.Console.Stream = ConsoleNo

	logRoll := LogWithConfig[MyLogStruct](cfg)
	for i := 0; i < 100000; i++ {
		logRoll.Info("log roll test,log roll test,log roll test,log roll test,log roll test,log roll test,log roll test,log roll test")
	}
}

func TestChangeLevel(t *testing.T) {
	log := Log[MyLogStruct]()
	log.Debug("debug hello 1")
	log.Info("info hello 1")
	log.Warn("warn hello 1")

	log.SetLevel(LevelDebug)
	log.Debug("debug hello 2")
	log.Info("info hello 2")
	log.Warn("warn hello 2")

	log.SetLevel(LevelWarn)
	log.Debug("debug hello 3")
	log.Info("info hello 3")
	log.Warn("warn hello 3")

	log.SetLevel(LevelInfo)
	log.Debug("debug hello 4")
	log.Info("info hello 4")
	log.Warn("warn hello 4")

}

func TestLogWire(t *testing.T) {
	myLog := factory.New[MyLog]()
	myLog.RunLog()
}
