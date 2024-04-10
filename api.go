package log

import (
	"github.com/expgo/config"
	"github.com/expgo/factory"
	"github.com/expgo/generic"
	"github.com/gobwas/glob"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"reflect"
	"time"
)

var logs generic.Cache[string, any]

func must[T any](log Logger[T], err error) Logger[T] {
	if err != nil {
		panic(err)
	}
	return log
}

type Logger[T any] interface {
	Level() Level
	SetLevel(lvl Level)
	TemporarySetLevel(lvl Level, d time.Duration)

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

func Log[T any]() Logger[T] {
	return must(New[T]())
}

func LogWithPath[T any](path string) Logger[T] {
	return must(NewWithPath[T](path))
}

func LogWithConfig[T any](cfg *Config) Logger[T] {
	return must(NewWithConfig[T](cfg))
}

func Lazy[T any]() {
	factory.Interface[Logger[T]]().SetInitFunc(func() Logger[T] {
		return Log[T]()
	})
}

func LazyWithPath[T any](path string) {
	factory.Interface[Logger[T]]().SetInitFunc(func() Logger[T] {
		return LogWithPath[T](path)
	})
}

func LasyWithConfig[T any](cfg *Config) {
	factory.Interface[Logger[T]]().SetInitFunc(func() Logger[T] {
		return LogWithConfig[T](cfg)
	})
}

func New[T any]() (Logger[T], error) {
	return NewWithPath[T]("logging")
}

func NewWithPath[T any](path string) (Logger[T], error) {
	cfg, err := config.New[Config](path)
	if err != nil {
		return nil, err
	}
	return NewWithConfig[T](cfg)
}

func NewWithConfig[T any](cfg *Config) (Logger[T], error) {
	vt := reflect.TypeOf((*T)(nil)).Elem()
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	typePath := vt.PkgPath() + "/" + vt.Name()

	log, err := logs.GetOrLoad(typePath, func(k string) (any, error) {
		l := &logger[T]{}
		l.level = zap.NewAtomicLevel()

		ec := zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		cores := []zapcore.Core{}
		l.SetLevel(cfg.GetZapLevelByType(typePath))

		if cfg.Console.Stream != ConsoleNo {

			if cfg.Console.Encoder == EncoderText {
				ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
			} else {
				ec.EncodeLevel = zapcore.CapitalLevelEncoder
			}

			consoleWriter := zapcore.Lock(os.Stdout)
			if cfg.Console.Stream == ConsoleStderr {
				consoleWriter = zapcore.Lock(os.Stderr)
			}

			consoleEncoder := zapcore.NewConsoleEncoder(ec)
			if cfg.Console.Encoder == EncoderJson {
				consoleEncoder = zapcore.NewJSONEncoder(ec)
			}

			consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, &l.level)
			cores = append(cores, consoleCore)
		}

		if len(cfg.File.Filename) > 0 {
			ec.EncodeLevel = zapcore.CapitalLevelEncoder

			fileWriter := zapcore.AddSync(&lumberjack.Logger{
				Filename:   cfg.File.Filename,
				MaxSize:    cfg.File.MaxSize,
				MaxAge:     cfg.File.MaxAge,
				MaxBackups: cfg.File.MaxBackups,
				LocalTime:  cfg.File.LocalTime,
				Compress:   cfg.File.Compress,
			})

			fileEncoder := zapcore.NewJSONEncoder(ec)
			if cfg.File.Encoder == EncoderText {
				fileEncoder = zapcore.NewConsoleEncoder(ec)
			}

			fileCore := zapcore.NewCore(fileEncoder, fileWriter, &l.level)
			cores = append(cores, fileCore)
		}

		l.base = zap.New(zapcore.NewTee(cores...)).Named(k)

		return l, nil
	})

	if err != nil {
		return nil, err
	}

	return log.(Logger[T]), nil
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
