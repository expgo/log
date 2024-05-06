package log

import (
	"github.com/expgo/config"
	"github.com/expgo/factory"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	logs = map[string]Logger{}

	log := Log[MyLogStruct]()
	log.Info("hello")
}

func TestLevel(t *testing.T) {
	logs = map[string]Logger{}

	cfg, _ := config.New[Config]("")
	cfg.Level["*Struct"] = LevelDebug
	cfg.File.Filename = "log/app.log"

	structLog := LogWithConfig[MyLogStruct](cfg)
	log := Log[MyLog]()

	structLogMsgs := []string{}
	structLog.AddHook(func(level Level, t time.Time, name string, msg string) {
		structLogMsgs = append(structLogMsgs, msg)
	})

	logMsgs := []string{}
	log.AddHook(func(level Level, t time.Time, name string, msg string) {
		logMsgs = append(logMsgs, msg)
	})

	structLog.Debug("debug hello")
	log.Debug("debug hello")

	structLog.Info("info hello")
	log.Info("info hello")

	structExpectMsgs := []string{"debug hello", "info hello"}

	logExpectMsgs := []string{"info hello"}

	assert.Equal(t, structExpectMsgs, structLogMsgs)
	assert.Equal(t, logExpectMsgs, logMsgs)
}

func TestLogRoll(t *testing.T) {
	logs = map[string]Logger{}

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
	logs = map[string]Logger{}

	log := Log[MyLogStruct]()

	msgs := []string{}
	log.AddHook(func(level Level, t time.Time, name string, msg string) {
		msgs = append(msgs, msg)
	})
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

	expectMsgs := []string{
		"info hello 1", "warn hello 1", "debug hello 2", "info hello 2",
		"warn hello 2", "warn hello 3", "info hello 4", "warn hello 4",
	}

	assert.Equal(t, expectMsgs, msgs)
}

func TestLogWire(t *testing.T) {
	logs = map[string]Logger{}

	myLog := factory.New[MyLog]()
	msgs := []string{}
	myLog.log.AddHook(func(level Level, t time.Time, name string, msg string) {
		msgs = append(msgs, msg)
	})
	myLog.RunLog()
	expectMsgs := []string{"info", "warn", "error"}
	assert.Equal(t, expectMsgs, msgs)
}
