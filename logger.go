package log

import (
	"fmt"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
	_multipleErrMsg     = "Multiple errors without a key."
)

type Logger[T any] struct {
	base *zap.Logger
}

// Named adds a sub-scope to the logger's name. See Logger.Named for details.
//func (s *logger[T]) Named(name string) *Logger[T] {
//	return &logger{base: s.base.Named(name)}
//}

// WithOptions clones the current Logger, applies the supplied Options,
// and returns the result. It's safe to use concurrently.
//func (s *logger[T]) WithOptions(opts ...zap.Option) *Logger[T] {
//	base := s.base
//	for _, opt := range opts {
//		opt.apply(base)
//	}
//	return &logger[T]{base: base}
//}

// With adds a variadic number of fields to the logging context. It accepts a
// mix of strongly-typed Field objects and loosely-typed key-value pairs. When
// processing pairs, the first element of the pair is used as the field key
// and the second as the field value.
//
// For example,
//
//	 sugaredLogger.With(
//	   "hello", "world",
//	   "failure", errors.New("oh no"),
//	   Stack(),
//	   "count", 42,
//	   "user", User{Name: "alice"},
//	)
//
// is the equivalent of
//
//	unsugared.With(
//	  String("hello", "world"),
//	  String("failure", "oh no"),
//	  Stack(),
//	  Int("count", 42),
//	  Object("user", User{Name: "alice"}),
//	)
//
// Note that the keys in key-value pairs should be strings. In development,
// passing a non-string key panics. In production, the logger is more
// forgiving: a separate error is logged, but the key-value pair is skipped
// and execution continues. Passing an orphaned key triggers similar behavior:
// panics in development and errors in production.
//func (s *logger[T]) With(args ...interface{}) *Logger[T] {
//	return &logger[T]{base: s.base.With(s.sweetenFields(args)...)}
//}

// WithLazy adds a variadic number of fields to the logging context lazily.
// The fields are evaluated only if the logger is further chained with [With]
// or is written to with any of the log level methods.
// Until that occurs, the logger may retain references to objects inside the fields,
// and logging will reflect the state of an object at the time of logging,
// not the time of WithLazy().
//
// Similar to [With], fields added to the child don't affect the parent,
// and vice versa. Also, the keys in key-value pairs should be strings. In development,
// passing a non-string key panics, while in production it logs an error and skips the pair.
// Passing an orphaned key has the same behavior.
//func (s *logger[T]) WithLazy(args ...interface{}) *Logger[T] {
//	return &logger[T]{base: s.base.WithLazy(s.sweetenFields(args)...)}
//}

// Level reports the minimum enabled level for this logger.
//
// For NopLoggers, this is [zapcore.InvalidLevel].
//func (s *logger[T]) Level() zapcore.Level {
//	return zapcore.LevelOf(s.base.core)
//}

// Log logs the provided arguments at provided level.
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) Log(lvl zapcore.Level, args ...interface{}) {
	s.log(lvl, "", args, nil)
}

// Debug logs the provided arguments at [DebugLevel].
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) Debug(args ...interface{}) {
	s.log(zapcore.DebugLevel, "", args, nil)
}

// Info logs the provided arguments at [InfoLevel].
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) Info(args ...interface{}) {
	s.log(zapcore.InfoLevel, "", args, nil)
}

// Warn logs the provided arguments at [WarnLevel].
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) Warn(args ...interface{}) {
	s.log(zapcore.WarnLevel, "", args, nil)
}

// Error logs the provided arguments at [ErrorLevel].
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) Error(args ...interface{}) {
	s.log(zapcore.ErrorLevel, "", args, nil)
}

// DPanic logs the provided arguments at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) DPanic(args ...interface{}) {
	s.log(zapcore.DPanicLevel, "", args, nil)
}

// Panic constructs a message with the provided arguments and panics.
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) Panic(args ...interface{}) {
	s.log(zapcore.PanicLevel, "", args, nil)
}

// Fatal constructs a message with the provided arguments and calls os.Exit.
// Spaces are added between arguments when neither is a string.
func (s *Logger[T]) Fatal(args ...interface{}) {
	s.log(zapcore.FatalLevel, "", args, nil)
}

// Logf formats the message according to the format specifier
// and logs it at provided level.
func (s *Logger[T]) Logf(lvl zapcore.Level, template string, args ...interface{}) {
	s.log(lvl, template, args, nil)
}

// Debugf formats the message according to the format specifier
// and logs it at [DebugLevel].
func (s *Logger[T]) Debugf(template string, args ...interface{}) {
	s.log(zapcore.DebugLevel, template, args, nil)
}

// Infof formats the message according to the format specifier
// and logs it at [InfoLevel].
func (s *Logger[T]) Infof(template string, args ...interface{}) {
	s.log(zapcore.InfoLevel, template, args, nil)
}

// Warnf formats the message according to the format specifier
// and logs it at [WarnLevel].
func (s *Logger[T]) Warnf(template string, args ...interface{}) {
	s.log(zapcore.WarnLevel, template, args, nil)
}

// Errorf formats the message according to the format specifier
// and logs it at [ErrorLevel].
func (s *Logger[T]) Errorf(template string, args ...interface{}) {
	s.log(zapcore.ErrorLevel, template, args, nil)
}

// DPanicf formats the message according to the format specifier
// and logs it at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
func (s *Logger[T]) DPanicf(template string, args ...interface{}) {
	s.log(zapcore.DPanicLevel, template, args, nil)
}

// Panicf formats the message according to the format specifier
// and panics.
func (s *Logger[T]) Panicf(template string, args ...interface{}) {
	s.log(zapcore.PanicLevel, template, args, nil)
}

// Fatalf formats the message according to the format specifier
// and calls os.Exit.
func (s *Logger[T]) Fatalf(template string, args ...interface{}) {
	s.log(zapcore.FatalLevel, template, args, nil)
}

// Logw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *Logger[T]) Logw(lvl zapcore.Level, msg string, keysAndValues ...interface{}) {
	s.log(lvl, msg, nil, keysAndValues)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func (s *Logger[T]) Debugw(msg string, keysAndValues ...interface{}) {
	s.log(zapcore.DebugLevel, msg, nil, keysAndValues)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *Logger[T]) Infow(msg string, keysAndValues ...interface{}) {
	s.log(zapcore.InfoLevel, msg, nil, keysAndValues)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *Logger[T]) Warnw(msg string, keysAndValues ...interface{}) {
	s.log(zapcore.WarnLevel, msg, nil, keysAndValues)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *Logger[T]) Errorw(msg string, keysAndValues ...interface{}) {
	s.log(zapcore.ErrorLevel, msg, nil, keysAndValues)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (s *Logger[T]) DPanicw(msg string, keysAndValues ...interface{}) {
	s.log(zapcore.DPanicLevel, msg, nil, keysAndValues)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (s *Logger[T]) Panicw(msg string, keysAndValues ...interface{}) {
	s.log(zapcore.PanicLevel, msg, nil, keysAndValues)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (s *Logger[T]) Fatalw(msg string, keysAndValues ...interface{}) {
	s.log(zapcore.FatalLevel, msg, nil, keysAndValues)
}

// Logln logs a message at provided level.
// Spaces are always added between arguments.
func (s *Logger[T]) Logln(lvl zapcore.Level, args ...interface{}) {
	s.logln(lvl, args, nil)
}

// Debugln logs a message at [DebugLevel].
// Spaces are always added between arguments.
func (s *Logger[T]) Debugln(args ...interface{}) {
	s.logln(zapcore.DebugLevel, args, nil)
}

// Infoln logs a message at [InfoLevel].
// Spaces are always added between arguments.
func (s *Logger[T]) Infoln(args ...interface{}) {
	s.logln(zapcore.InfoLevel, args, nil)
}

// Warnln logs a message at [WarnLevel].
// Spaces are always added between arguments.
func (s *Logger[T]) Warnln(args ...interface{}) {
	s.logln(zapcore.WarnLevel, args, nil)
}

// Errorln logs a message at [ErrorLevel].
// Spaces are always added between arguments.
func (s *Logger[T]) Errorln(args ...interface{}) {
	s.logln(zapcore.ErrorLevel, args, nil)
}

// DPanicln logs a message at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are always added between arguments.
func (s *Logger[T]) DPanicln(args ...interface{}) {
	s.logln(zapcore.DPanicLevel, args, nil)
}

// Panicln logs a message at [PanicLevel] and panics.
// Spaces are always added between arguments.
func (s *Logger[T]) Panicln(args ...interface{}) {
	s.logln(zapcore.PanicLevel, args, nil)
}

// Fatalln logs a message at [FatalLevel] and calls os.Exit.
// Spaces are always added between arguments.
func (s *Logger[T]) Fatalln(args ...interface{}) {
	s.logln(zapcore.FatalLevel, args, nil)
}

// Sync flushes any buffered log entries.
func (s *Logger[T]) Sync() error {
	return s.base.Sync()
}

// log message with Sprint, Sprintf, or neither.
func (s *Logger[T]) log(lvl zapcore.Level, template string, fmtArgs []interface{}, context []interface{}) {
	// If logging at this level is completely disabled, skip the overhead of
	// string formatting.
	if lvl < zap.DPanicLevel && !s.base.Core().Enabled(lvl) {
		return
	}

	msg := getMessage(template, fmtArgs)
	if ce := s.base.Check(lvl, msg); ce != nil {
		ce.Write(s.sweetenFields(context)...)
	}
}

// logln message with Sprintln
func (s *Logger[T]) logln(lvl zapcore.Level, fmtArgs []interface{}, context []interface{}) {
	if lvl < zap.DPanicLevel && !s.base.Core().Enabled(lvl) {
		return
	}

	msg := getMessageln(fmtArgs)
	if ce := s.base.Check(lvl, msg); ce != nil {
		ce.Write(s.sweetenFields(context)...)
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

func (s *Logger[T]) sweetenFields(args []interface{}) []zap.Field {
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
				s.base.Error(_multipleErrMsg, zap.Error(err))
			}
			i++
			continue
		}

		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			s.base.Error(_oddNumberErrMsg, zap.Any("ignored", args[i]))
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
		s.base.Error(_nonStringKeyErrMsg, zap.Array("invalid", invalid))
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
