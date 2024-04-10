package example

import (
	"github.com/expgo/factory"
	"testing"
)

func TestLogWithWire(t *testing.T) {
	myLog := factory.New[MyLog]()
	myLog.WriteLog()
}
