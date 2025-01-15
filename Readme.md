# config file controlled logging system for golang

Build a logging system that can be controlled by a config file.

## config file

logging:                 # default logging config name
  level:
    "*": debug           # all loggers will be set to debug level
    "*MyLog": warn       # only MyLog will be set to warn level
  console:
    stream: stdout       # console output stream, will be `no`, `stdout` or `stderr`, default is `stdout`. `no` means no console output.
    encoder: text        # encoder type, will be `text` or `json`, default is `text`.
  file:
    filename: log/a.log  # log file name, Backup log files will be retained in the same directory
    encoder: text        # log file encoder
    maxsize: 100         # log file max size, the maximum size in megabytes of the log file before it gets rotated, It defaults to 100 megabytes.
    maxage: 30           # log file max age, the maximum number of days to retain old log files based on the timestamp encoded in their filename. Default is 30 days.
    maxbackups: 30       # log file max backups, the maximum number of old log files to retain. Default is 30.
    compress: true       # determines if the rotated log files should be compressed using gzip. Default is true.
    encoder: text        # log file encoder, will be `text` or `json`, default is `text`.
  withcaller: true       # configures the Logger to annotate each message with the filename, line number, and function name of caller. Default is true.
  withlogname: short     # If log the file type name. will be `short`, `full` or `none`. Default is `short`. `short` means the file name without the directory path. `full` means the file name with the directory path. `none` means no file name.

## how to use custom level name
- If no config file, there will be a default `"*": info` under the `level` section.
- The level of the key is the structure full type with path. For example, a struct `MyLog` `github.com/mind/log/v2/logger.go` file, the full key will be `"github.com/mind/log/v2.MyLog": debug`.

## all the level name
- debug
- info
- warn
- error
- dpanic
- panic
- fatal
- invalid
