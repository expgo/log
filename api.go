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

		cores := []zapcore.Core{}
		if cfg.Console != ConsoleNo {
			consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
			consoleWriter := zapcore.Lock(os.Stdout)
			if cfg.Console == ConsoleStderr {
				consoleWriter = zapcore.Lock(os.Stderr)
			}
			consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, cfg.GetZapLevelByType(typePath))
			cores = append(cores, consoleCore)
		}

		if len(cfg.File.Filename) > 0 {
			fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
			fileWriter := zapcore.AddSync(&lumberjack.Logger{
				Filename:   cfg.File.Filename,
				MaxSize:    cfg.File.MaxSize,
				MaxAge:     cfg.File.MaxAge,
				MaxBackups: cfg.File.MaxBackups,
				LocalTime:  cfg.File.LocalTime,
				Compress:   cfg.File.Compress,
			})
			fileCore := zapcore.NewCore(fileEncoder, fileWriter, cfg.GetZapLevelByType(typePath))
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
