package example

import (
	"github.com/expgo/config"
	"github.com/expgo/factory"
	"github.com/expgo/log"
	"testing"
	"time"
)

func TestLogWithWire(t *testing.T) {
	config.SetDefaultFilename("log.yml")

	myLog := factory.New[MyLog]()
	myLog.WriteLog()

	log.TemporarySetLevel("*MyLog", log.LevelDebug, 3*time.Second)

	time.Sleep(2 * time.Second)

	log.TemporarySetLevel("*MyLog", log.LevelDebug, 3*time.Second)

	for i := 0; i < 10; i++ {
		myLog.WriteLog()
		time.Sleep(1 * time.Second)
	}

	myLog1 := factory.New[MyLog1]()
	myLog1.WriteLog()

	logger.Debug("debug 123")
	logger.Info("info 123")
	logger.Warn("warn 123")
	logger.Error("err 123")
}
