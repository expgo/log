package log

import (
	"github.com/expgo/generic"
	"github.com/gobwas/glob"
	"io"
	"reflect"
	"time"
)

var logs generic.Cache[string, Logger]

const DefaultConfigPath = "logging"

func Must(log Logger, err error) Logger {
	if err != nil {
		panic(err)
	}
	return log
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
	return LogWithConfigPath[T](DefaultConfigPath)
}

func LogWithConfigPath[T any](cfgPath string) Logger {
	return Must(NewWithConfigPath(new(T), cfgPath))
}

func LogWithConfig[T any](cfg *Config) Logger {
	vt := reflect.TypeOf((*T)(nil))
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	typePath := vt.PkgPath() + "." + vt.Name()

	return Must(NewWithTypePathAndConfig(typePath, cfg))
}

func New(t any) (Logger, error) {
	return NewWithConfigPath(t, DefaultConfigPath)
}

func NewWithConfigPath(t any, cfgPath string) (Logger, error) {
	vt := reflect.TypeOf(t)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	typePath := vt.PkgPath() + "." + vt.Name()

	return NewWithTypePathAndConfigPath(typePath, cfgPath)
}

// @Factory(params={self, "value:logging"})
func FactoryWithConfigPath(t any, cfgPath string) Logger {
	return Must(NewWithConfigPath(t, cfgPath))
}

func FactoryWithTypePathAndConfigPath(typePath string, cfgPath string) Logger {
	return Must(NewWithTypePathAndConfigPath(typePath, cfgPath))
}

func NewWithTypePathAndConfigPath(typePath string, cfgPath string) (Logger, error) {
	log, err := logs.GetOrLoad(typePath, func(k string) (Logger, error) {
		l := &logger{
			typePath: typePath,
			cfgPath:  cfgPath,
		}

		return l, nil
	})

	if err != nil {
		return nil, err
	}

	return log, nil
}

func NewWithTypePathAndConfig(typePath string, cfg *Config) (Logger, error) {
	log, err := logs.GetOrLoad(typePath, func(k string) (Logger, error) {
		l := &logger{
			typePath: typePath,
			cfg:      cfg,
		}

		return l, nil
	})

	if err != nil {
		return nil, err
	}

	return log, nil
}

func SetLevel(logPathGlob string, level Level) {
	TemporarySetLevel(logPathGlob, level, 0)
}

func TemporarySetLevel(logPath string, level Level, d time.Duration) {
	logPathGlob := glob.MustCompile(logPath)
	keys := logs.Keys()

	for _, key := range keys {
		if logPathGlob.Match(key) {
			l, loaded := logs.Get(key)
			if loaded {
				if log, ok := l.(ITemporarySetLevel); ok {
					log.TemporarySetLevel(level, d)
				}
			}
		}
	}
}

func Sync() error {
	for _, key := range logs.Keys() {
		if l, ok := logs.Get(key); ok {
			if err := l.Sync(); err != nil {
				return err
			}
		}
	}

	return nil
}
