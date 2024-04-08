package log

import "go.uber.org/zap/zapcore"

type LogFile struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" yaml:"maxsize" value:"100"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage" yaml:"maxage" value:"30"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups" yaml:"maxbackups" value:"30"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `json:"localtime" yaml:"localtime"  value:"false"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"  value:"true"`
	// contains filtered or unexported fields
}

/*
Console is en enum

	@Enum {
		no
		stdout
		stderr
	}
*/
type Console string

/*
Level is an enum

	@EnumConfig(PanicIfInvalid)
	@Enum {
		Debug = -1
		Info
		Warn
		Error
		DPanic
		Panic
		Fatal
	}
*/
type Level int8

func (l Level) ToZapLevel() zapcore.Level {
	return zapcore.Level(l)
}

type Config struct {
	Level map[string]Level
	// Console is output to
	Console Console `value:"no"`
	File    LogFile
}

func (c *Config) Init() {
	c.Level = make(map[string]Level)
	c.Level["*"] = LevelInfo
}
