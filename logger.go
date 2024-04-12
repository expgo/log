package log

import (
	"fmt"
	"github.com/expgo/config"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
	"time"
)

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
	_multipleErrMsg     = "Multiple errors without a key."
)

type logger struct {
	base        *zap.Logger
	level       zap.AtomicLevel
	originLevel Level
	tempTimer   *time.Timer
	timerLock   sync.Mutex

	typePath string
	cfgPath  string
	cfg      *Config
	once     sync.Once
}

type ITemporarySetLevel interface {
	TemporarySetLevel(level Level, d time.Duration)
}

func fullCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", caller.Function) + caller.TrimmedPath())
}

func (l *logger) init() {
	l.once.Do(func() {
		if l.cfg == nil {
			if len(l.cfgPath) == 0 {
				panic("cfgPath or cfg must set one")
			}

			cfg, err := config.New[Config](l.cfgPath)
			if err != nil {
				panic(err)
			}

			l.cfg = cfg
		}

		cfg := l.cfg
		l.level = zap.NewAtomicLevelAt(cfg.GetZapLevelByType(l.typePath).ToZapLevel())

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
			EncodeCaller:   fullCallerEncoder,
		}

		cores := []zapcore.Core{}

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

		name := cfg.GetName(l.typePath)
		if len(name) > 0 {
			l.base = l.base.Named(name)
		}

		options := []zap.Option{}
		options = append(options, zap.AddCallerSkip(2))
		options = append(options, zap.WithCaller(cfg.WithCaller))

		l.base = l.base.WithOptions(options...)
	})
}

// Level reports the minimum enabled level for this logger.
func (l *logger) Level() Level {
	l.init()

	result := Level(l.level.Level())
	if result.IsValid() {
		return result
	} else {
		return LevelInvalid
	}
}

// SetLevel set the log's level
func (l *logger) SetLevel(level Level) {
	l.level.SetLevel(level.ToZapLevel())
}

func (l *logger) rollbackLevel() {
	l.timerLock.Lock()
	defer l.timerLock.Unlock()

	if l.tempTimer != nil {
		l.tempTimer.Stop()
		l.tempTimer = nil
	}

	l.SetLevel(l.originLevel)
}

func (l *logger) TemporarySetLevel(level Level, d time.Duration) {
	l.timerLock.Lock()
	defer l.timerLock.Unlock()

	if l.tempTimer == nil {
		l.originLevel = l.Level()
	} else {
		l.tempTimer.Stop()
		l.tempTimer = nil
	}

	if d > 0 {
		l.tempTimer = time.AfterFunc(d, func() {
			l.rollbackLevel()
		})
	}

	l.SetLevel(level)
}

// Log logs the provided arguments at provided level.
// Spaces are added between arguments when neither is a string.
func (l *logger) Log(lvl Level, args ...interface{}) {
	l.log(lvl.ToZapLevel(), "", args, nil)
}

// Debug logs the provided arguments at [DebugLevel].
// Spaces are added between arguments when neither is a string.
func (l *logger) Debug(args ...interface{}) {
	l.log(zapcore.DebugLevel, "", args, nil)
}

// Info logs the provided arguments at [InfoLevel].
// Spaces are added between arguments when neither is a string.
func (l *logger) Info(args ...interface{}) {
	l.log(zapcore.InfoLevel, "", args, nil)
}

// Warn logs the provided arguments at [WarnLevel].
// Spaces are added between arguments when neither is a string.
func (l *logger) Warn(args ...interface{}) {
	l.log(zapcore.WarnLevel, "", args, nil)
}

// Error logs the provided arguments at [ErrorLevel].
// Spaces are added between arguments when neither is a string.
func (l *logger) Error(args ...interface{}) {
	l.log(zapcore.ErrorLevel, "", args, nil)
}

// DPanic logs the provided arguments at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are added between arguments when neither is a string.
func (l *logger) DPanic(args ...interface{}) {
	l.log(zapcore.DPanicLevel, "", args, nil)
}

// Panic constructs a message with the provided arguments and panics.
// Spaces are added between arguments when neither is a string.
func (l *logger) Panic(args ...interface{}) {
	l.log(zapcore.PanicLevel, "", args, nil)
}

// Fatal constructs a message with the provided arguments and calls os.Exit.
// Spaces are added between arguments when neither is a string.
func (l *logger) Fatal(args ...interface{}) {
	l.log(zapcore.FatalLevel, "", args, nil)
}

// Logf formats the message according to the format specifier
// and logs it at provided level.
func (l *logger) Logf(lvl Level, template string, args ...interface{}) {
	l.log(lvl.ToZapLevel(), template, args, nil)
}

// Debugf formats the message according to the format specifier
// and logs it at [DebugLevel].
func (l *logger) Debugf(template string, args ...interface{}) {
	l.log(zapcore.DebugLevel, template, args, nil)
}

// Infof formats the message according to the format specifier
// and logs it at [InfoLevel].
func (l *logger) Infof(template string, args ...interface{}) {
	l.log(zapcore.InfoLevel, template, args, nil)
}

// Warnf formats the message according to the format specifier
// and logs it at [WarnLevel].
func (l *logger) Warnf(template string, args ...interface{}) {
	l.log(zapcore.WarnLevel, template, args, nil)
}

// Errorf formats the message according to the format specifier
// and logs it at [ErrorLevel].
func (l *logger) Errorf(template string, args ...interface{}) {
	l.log(zapcore.ErrorLevel, template, args, nil)
}

// DPanicf formats the message according to the format specifier
// and logs it at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
func (l *logger) DPanicf(template string, args ...interface{}) {
	l.log(zapcore.DPanicLevel, template, args, nil)
}

// Panicf formats the message according to the format specifier
// and panics.
func (l *logger) Panicf(template string, args ...interface{}) {
	l.log(zapcore.PanicLevel, template, args, nil)
}

// Fatalf formats the message according to the format specifier
// and calls os.Exit.
func (l *logger) Fatalf(template string, args ...interface{}) {
	l.log(zapcore.FatalLevel, template, args, nil)
}

// Logw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *logger) Logw(lvl Level, msg string, keysAndValues ...interface{}) {
	l.log(lvl.ToZapLevel(), msg, nil, keysAndValues)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func (l *logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.log(zapcore.DebugLevel, msg, nil, keysAndValues)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *logger) Infow(msg string, keysAndValues ...interface{}) {
	l.log(zapcore.InfoLevel, msg, nil, keysAndValues)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.log(zapcore.WarnLevel, msg, nil, keysAndValues)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.log(zapcore.ErrorLevel, msg, nil, keysAndValues)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (l *logger) DPanicw(msg string, keysAndValues ...interface{}) {
	l.log(zapcore.DPanicLevel, msg, nil, keysAndValues)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (l *logger) Panicw(msg string, keysAndValues ...interface{}) {
	l.log(zapcore.PanicLevel, msg, nil, keysAndValues)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (l *logger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.log(zapcore.FatalLevel, msg, nil, keysAndValues)
}

// Logln logs a message at provided level.
// Spaces are always added between arguments.
func (l *logger) Logln(lvl zapcore.Level, args ...interface{}) {
	l.logln(lvl, args, nil)
}

// Debugln logs a message at [DebugLevel].
// Spaces are always added between arguments.
func (l *logger) Debugln(args ...interface{}) {
	l.logln(zapcore.DebugLevel, args, nil)
}

// Infoln logs a message at [InfoLevel].
// Spaces are always added between arguments.
func (l *logger) Infoln(args ...interface{}) {
	l.logln(zapcore.InfoLevel, args, nil)
}

// Warnln logs a message at [WarnLevel].
// Spaces are always added between arguments.
func (l *logger) Warnln(args ...interface{}) {
	l.logln(zapcore.WarnLevel, args, nil)
}

// Errorln logs a message at [ErrorLevel].
// Spaces are always added between arguments.
func (l *logger) Errorln(args ...interface{}) {
	l.logln(zapcore.ErrorLevel, args, nil)
}

// DPanicln logs a message at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are always added between arguments.
func (l *logger) DPanicln(args ...interface{}) {
	l.logln(zapcore.DPanicLevel, args, nil)
}

// Panicln logs a message at [PanicLevel] and panics.
// Spaces are always added between arguments.
func (l *logger) Panicln(args ...interface{}) {
	l.logln(zapcore.PanicLevel, args, nil)
}

// Fatalln logs a message at [FatalLevel] and calls os.Exit.
// Spaces are always added between arguments.
func (l *logger) Fatalln(args ...interface{}) {
	l.logln(zapcore.FatalLevel, args, nil)
}

// Sync flushes any buffered log entries.
func (l *logger) Sync() error {
	l.init()
	return l.base.Sync()
}

// log message with Sprint, Sprintf, or neither.
func (l *logger) log(lvl zapcore.Level, template string, fmtArgs []interface{}, context []interface{}) {
	l.init()
	// If logging at this level is completely disabled, skip the overhead of
	// string formatting.
	if lvl < zap.DPanicLevel && !l.base.Core().Enabled(lvl) {
		return
	}

	msg := getMessage(template, fmtArgs)
	if ce := l.base.Check(lvl, msg); ce != nil {
		ce.Write(l.sweetenFields(context)...)
	}
}

// logln message with Sprintln
func (l *logger) logln(lvl zapcore.Level, fmtArgs []interface{}, context []interface{}) {
	l.init()
	if lvl < zap.DPanicLevel && !l.base.Core().Enabled(lvl) {
		return
	}

	msg := getMessageln(fmtArgs)
	if ce := l.base.Check(lvl, msg); ce != nil {
		ce.Write(l.sweetenFields(context)...)
	}
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

// getMessageln format with Sprintln.
func getMessageln(fmtArgs []interface{}) string {
	msg := fmt.Sprintln(fmtArgs...)
	return msg[:len(msg)-1]
}

func (l *logger) sweetenFields(args []interface{}) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	var (
		// Allocate enough space for the worst case; if users pass only structured
		// fields, we shouldn't penalize them with extra allocations.
		fields    = make([]zap.Field, 0, len(args))
		invalid   invalidPairs
		seenError bool
	)

	for i := 0; i < len(args); {
		// This is a strongly-typed field. Consume it and move on.
		if f, ok := args[i].(zap.Field); ok {
			fields = append(fields, f)
			i++
			continue
		}

		// If it is an error, consume it and move on.
		if err, ok := args[i].(error); ok {
			if !seenError {
				seenError = true
				fields = append(fields, zap.Error(err))
			} else {
				l.base.Error(_multipleErrMsg, zap.Error(err))
			}
			i++
			continue
		}

		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			l.base.Error(_oddNumberErrMsg, zap.Any("ignored", args[i]))
			break
		}

		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			if cap(invalid) == 0 {
				invalid = make(invalidPairs, 0, len(args)/2)
			}
			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			fields = append(fields, zap.Any(keyStr, val))
		}
		i += 2
	}

	// If we encountered any invalid key-value pairs, log an error.
	if len(invalid) > 0 {
		l.base.Error(_nonStringKeyErrMsg, zap.Array("invalid", invalid))
	}
	return fields
}

type invalidPair struct {
	position   int
	key, value interface{}
}

func (p invalidPair) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("position", int64(p.position))
	zap.Any("key", p.key).AddTo(enc)
	zap.Any("value", p.value).AddTo(enc)
	return nil
}

type invalidPairs []invalidPair

func (ps invalidPairs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	var err error
	for i := range ps {
		err = multierr.Append(err, enc.AppendObject(ps[i]))
	}
	return err
}
