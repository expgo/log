package log

import (
	"github.com/expgo/structure"
	"github.com/expgo/sync"
	"github.com/gobwas/glob"
	"io"
	"reflect"
	"time"
)

var logs = map[string]Logger{}
var logsLock = sync.NewRWMutex()

const DefaultConfigPath = "logging"

// InnerLog inner log struct
type InnerLog struct {
	L Logger `new:""`
}

type Logger interface {
	Level() Level
	SetLevel(lvl Level)
	TemporarySetLevel(lvl Level, d time.Duration)
	AddHook(func(level Level, t time.Time, name string, msg string))
	Writer() io.Writer
	Sync() error

	Log(lvl Level, args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Panic(args ...any)
	Fatal(args ...any)

	Logf(lvl Level, template string, args ...any)
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Panicf(template string, args ...any)
	Fatalf(template string, args ...any)

	Logw(lvl Level, msg string, keysAndValues ...any)
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Panicw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
}

func Log[T any]() Logger {
	return getOrNewLog(new(T), DefaultConfigPath, nil)
}

func LogWithConfigPath[T any](cfgPath string) Logger {
	return getOrNewLog(new(T), cfgPath, nil)
}

func LogWithConfig[T any](cfg *Config) Logger {
	return getOrNewLog(new(T), "", cfg)
}

func New(t any) Logger {
	return getOrNewLog(t, DefaultConfigPath, nil)
}

// @Factory(params={self, "value:logging"})
func NewWithConfigPath(t any, cfgPath string) Logger {
	return getOrNewLog(t, cfgPath, nil)
}

func NewWithTypePathAndConfigPath(typePath string, cfgPath string) Logger {
	return getOrNewLogByPath(typePath, cfgPath, nil)
}

func NewWithTypePathAndConfig(typePath string, cfg *Config) Logger {
	return getOrNewLogByPath(typePath, "", cfg)
}

func getOrNewLog(t any, cfgPath string, cfg *Config) Logger {
	vt := reflect.TypeOf(t)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	typePath := vt.PkgPath() + "." + vt.Name()

	return getOrNewLogByPath(typePath, cfgPath, cfg)
}

func getOrNewLogByPath(typePath string, cfgPath string, cfg *Config) Logger {
	logsLock.RLock()
	log, ok := logs[typePath]
	logsLock.RUnlock()

	if !ok {
		log = &logger{
			typePath: typePath,
			cfgPath:  cfgPath,
			cfg:      cfg,
		}

		logsLock.Lock()
		logs[typePath] = log
		logsLock.Unlock()
	}

	return log
}

func SetLevel(logPathGlob string, level Level) {
	TemporarySetLevel(logPathGlob, level, 0)
}

func TemporarySetLevel(logPath string, level Level, d time.Duration) {
	logPathGlob := glob.MustCompile(logPath)

	logsLock.RLock()
	_logs := structure.CloneMap(logs)
	logsLock.RUnlock()

	for key, log := range _logs {
		if logPathGlob.Match(key) {
			log.TemporarySetLevel(level, d)
		}
	}
}

func Sync() error {
	logsLock.RLock()
	_logs := structure.CloneMap(logs)
	logsLock.RUnlock()

	for _, log := range _logs {
		if err := log.Sync(); err != nil {
			return err
		}
	}

	return nil
}
