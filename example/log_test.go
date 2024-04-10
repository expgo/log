package example

import (
	"github.com/expgo/factory"
	"github.com/expgo/log"
	"testing"
	"time"
)

func TestLogWithWire(t *testing.T) {
	myLog := factory.New[MyLog]()
	myLog.WriteLog()

	log.TemporarySetLevel("*MyLog", log.LevelDebug, 3*time.Second)

	time.Sleep(2 * time.Second)

	log.TemporarySetLevel("*MyLog", log.LevelDebug, 3*time.Second)

	for i := 0; i < 10; i++ {
		myLog.WriteLog()
		time.Sleep(1 * time.Second)
	}
}
