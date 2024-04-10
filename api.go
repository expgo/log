package log

import (
	"github.com/expgo/config"
	"github.com/expgo/generic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"reflect"
)

var logs generic.Cache[string, any]

func Log[T any]() *Logger[T] {
	log, err := New[T]()
	if err != nil {
		panic(err)
	}
	return log
}

func New[T any]() (*Logger[T], error) {
	cfg, err := config.New[Config]("logging")
	if err != nil {
		return nil, err
	}
	return NewWithConfig[T](cfg)
}

func NewWithConfig[T any](cfg *Config) (*Logger[T], error) {
	vt := reflect.TypeOf((*T)(nil)).Elem()
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	typePath := vt.PkgPath() + "/" + vt.Name()

	log, err := logs.GetOrLoad(typePath, func(k string) (any, error) {
		l := &Logger[T]{}
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

	return log.(*Logger[T]), nil
}
