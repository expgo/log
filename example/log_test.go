package example

import (
	"github.com/expgo/config"
	"github.com/expgo/factory"
	"github.com/expgo/log"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLogWithWire(t *testing.T) {
	config.SetDefaultFilename("log.yml")

	myLog := factory.New[MyLog]()
	msgs := []string{}
	myLog.log.AddHook(func(level log.Level, t time.Time, name string, msg string) {
		msgs = append(msgs, msg)
	})

	myLog.WriteLog()

	log.TemporarySetLevel("*MyLog", log.LevelDebug, 300*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	log.TemporarySetLevel("*MyLog", log.LevelDebug, 300*time.Millisecond)

	for i := 0; i < 10; i++ {
		myLog.WriteLog()
		time.Sleep(110 * time.Millisecond)
	}

	expMsgs := []string{
		"warn", "error", "debug", "info", "warn", "error", "debug", "info", "warn", "error", "debug", "info", "warn",
		"error", "warn", "error", "warn", "error", "warn", "error", "warn", "error", "warn", "error", "warn", "error",
		"warn", "error",
	}

	assert.Equal(t, expMsgs, msgs)
}

func TestLog1(t *testing.T) {
	config.SetDefaultFilename("log.yml")

	myLog1 := factory.New[MyLog1]()
	msgs := []string{}
	myLog1.log.AddHook(func(level log.Level, t time.Time, name string, msg string) {
		msgs = append(msgs, msg)
	})

	myLog1.WriteLog()

	expMsgs := []string{"debug 1", "info 2", "warn 3", "error 4"}
	assert.Equal(t, expMsgs, msgs)
}

func TestLogger(t *testing.T) {
	msgs := []string{}
	logger.AddHook(func(level log.Level, t time.Time, name string, msg string) {
		msgs = append(msgs, msg)
	})

	logger.Debug("debug 123")
	logger.Info("info 123")
	logger.Warn("warn 123")
	logger.Error("err 123")

	expMsgs := []string{"debug 123", "info 123", "warn 123", "err 123"}
	assert.Equal(t, expMsgs, msgs)
}
