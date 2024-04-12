package log

import (
	"fmt"
	"github.com/expgo/config"
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
	cfg, err := config.New[Config](cfgPath)
	if err != nil {
		return nil, err
	}

	vt := reflect.TypeOf(t)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	typePath := vt.PkgPath() + "." + vt.Name()

	return NewWithTypePathAndConfig(typePath, cfg)
}

// @Factory(params={self, "value:logging"})
func FactoryWithConfigPath(t any, cfgPath string) Logger {
	return Must(NewWithConfigPath(t, cfgPath))
}

func FactoryWithTypePathAndConfigPath(typePath string, cfgPath string) Logger {
	return Must(NewWithTypePathAndConfigPath(typePath, cfgPath))
}

func NewWithTypePathAndConfigPath(typePath string, cfgPath string) (Logger, error) {
	cfg, err := config.New[Config](cfgPath)
	if err != nil {
		return nil, err
	}
	return NewWithTypePathAndConfig(typePath, cfg)
}

func FullCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", caller.Function) + caller.TrimmedPath())
}

func NewWithTypePathAndConfig(typePath string, cfg *Config) (Logger, error) {
	log, err := logs.GetOrLoad(typePath, func(k string) (any, error) {
		l := &logger{}
		l.level = zap.NewAtomicLevel()

		ec := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   FullCallerEncoder,
		}

		cores := []zapcore.Core{}
		l.SetLevel(cfg.GetZapLevelByType(typePath))

		if cfg.Console.Stream != ConsoleNo {

			if cfg.Console.Encoder == EncoderText {
				ec.EncodeLevel = zapcore.LowercaseColorLevelEncoder
			} else {
				ec.EncodeLevel = zapcore.LowercaseLevelEncoder
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
			ec.EncodeLevel = zapcore.LowercaseLevelEncoder

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

		l.base = zap.New(zapcore.NewTee(cores...))

		name := cfg.GetName(typePath)
		if len(name) > 0 {
			l.base = l.base.Named(name)
		}

		options := []zap.Option{}
		options = append(options, zap.AddCallerSkip(2))
		options = append(options, zap.WithCaller(cfg.WithCaller))

		l.WithOptions(options...)

		return l, nil
	})

	if err != nil {
		return nil, err
	}

	return log.(Logger), nil
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
