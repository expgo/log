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
	cfg.File.Filename = "log/app.log"

	structLog, _ := NewWithConfig[MyLogStruct](cfg)
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

	logRoll, _ := NewWithConfig[MyLogStruct](cfg)
	for i := 0; i < 100000; i++ {
		logRoll.Info("log roll test,log roll test,log roll test,log roll test,log roll test,log roll test,log roll test,log roll test")
	}
}

func TestChangeLevel(t *testing.T) {
	log := Log[MyLogStruct]()
	log.Debug("debug hello 1")
	log.Info("info hello 1")

	log.SetLevel(LevelDebug.ToZapLevel())
	log.Debug("debug hello 2")
	log.Info("info hello 2")

	log.SetLevel(LevelWarn.ToZapLevel())
	log.Debug("debug hello 3")
	log.Info("info hello 3")

	log.SetLevel(LevelInfo.ToZapLevel())
	log.Debug("debug hello 4")
	log.Info("info hello 4")
}
