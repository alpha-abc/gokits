package logv1

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*terminal color format
"\x1b[0;%dm%s\x1b[0m"
*/
const (
	TerminalColorBlack = iota + 30
	TerminalColorRed
	TerminalColorGreen
	TerminalColorYellow
	TerminalColorBlue
	TerminalColorMagenta
	TerminalColorCyan
	TerminalColorWhite
)

/*log level*/
const (
	LevelDebug = 0
	LevelInfo  = 1
	LevelWarn  = 2
	LevelError = 3
	LevelFatal = 4
)

/*prefix*/
const (
	PrefixDebug = "[DEBUG]"
	PrefixInfo  = "[INFO ]"
	PrefixWarn  = "[WARN ]"
	PrefixError = "[ERROR]"
	PrefixFatal = "[FATAL]"
)

const (
	backupTypeName = "2006-01-02 15:04:05"
)

func init() {
	/**/
}

// Logger 日志实体
// 请传递正确的初始化值
// Outputs      example: []*os.File{os.Stdout}
// Level        example: LevelInfo
// TimeFormat   example: "2006-01-02 15:04:05.000"
// BackupType   example: "D" day, "size" size, "s" second, "m" minute, "h" hour, "M" month, "Y" year
// MaxSize      example: 1024 * 1024 * 1024 * 50 //50 MB
// CallDepth    example: 2 //固定值
// InitTime     example: time.Now()
type Logger struct {
	mux sync.Mutex

	outputs    []*os.File
	level      int
	timeFormat string
	backupType string
	maxSize    int64

	callDepth int
	initTime  time.Time
}

func (l *Logger) output(prefix, logstr string, level, color int) error {
	if l.level > level {
		return nil
	}

	var now = time.Now()

	l.mux.Lock()
	defer l.mux.Unlock()

	var buf []byte

	/*logstr format*/
	var tfmt = now.Format(l.timeFormat)

	var pc, fileName, lineNumber, ok = runtime.Caller(l.callDepth)
	var funcName = ""
	if !ok {
		return errors.New("runtime caller false")
	}

	for i := len(fileName) - 1; i > 0; i-- {
		if fileName[i] == '/' {
			fileName = fileName[i+1:]
			break
		}
	}

	funcName = runtime.FuncForPC(pc).Name()

	buf = append(buf, tfmt...)
	buf = append(buf, " "...)
	buf = append(buf, prefix...)

	buf = append(buf, " ["...)
	buf = append(buf, funcName...)
	buf = append(buf, "] "...)

	buf = append(buf, fileName...)
	buf = append(buf, ":"...)
	buf = append(buf, strconv.Itoa(lineNumber)...)

	buf = append(buf, " ▸ "...)

	buf = append(buf, logstr...)

	var _, err = l.write(&buf, now, color)
	if err != nil {
		return err
	}

	return nil
}

/*
 return, FALSE (index + 1, error), SUCCESS (0, nil)
*/
func (l *Logger) write(b *[]byte, time time.Time, color int) (int, error) {
	for i, f := range l.outputs {
		var fd = f.Fd()
		var name = f.Name()

		var finalBuf []byte
		if (fd == 1 && name == os.Stdout.Name()) || (fd == 2 && name == os.Stderr.Name()) {
			if TerminalColorBlack <= color && color <= TerminalColorWhite {
				finalBuf = append(finalBuf, "\x1b[0;"...)
				finalBuf = append(finalBuf, strconv.Itoa(color)...)
				finalBuf = append(finalBuf, "m"...)
				finalBuf = append(finalBuf, *b...)
				finalBuf = append(finalBuf, "\x1b[0m"...)
			} else {
				finalBuf = append(finalBuf, *b...)
			}
		} else {
			var hour, min, sec = time.Clock()
			var year, month, day = time.Date()

			var bakPath = ""

			var bakFmt = "%s.bak.%s"
			switch l.backupType {
			case "size":
				//file size
				var fileInfo, fiErr = f.Stat()
				if fiErr != nil {
					return i + 1, fiErr
				}

				if fileInfo.Size() > l.maxSize {
					bakPath = fmt.Sprintf(bakFmt, name, l.initTime.Format(backupTypeName))
					l.initTime = time
				}
			case "s":
				//second
				var _, _, lSec = l.initTime.Clock()
				if lSec < sec {
					bakPath = fmt.Sprintf(bakFmt, name, l.initTime.Format(backupTypeName))
					l.initTime = time
				}
			case "m":
				//minute
				var _, lMin, _ = l.initTime.Clock()
				if lMin < min {
					bakPath = fmt.Sprintf(bakFmt, name, l.initTime.Format(backupTypeName[0:len(backupTypeName)-3]))
					l.initTime = time
				}
			case "h":
				//hour
				var lHour, _, _ = l.initTime.Clock()
				if lHour < hour {
					bakPath = fmt.Sprintf(bakFmt, name, l.initTime.Format(backupTypeName[0:len(backupTypeName)-6]))
					l.initTime = time
				}
			case "D":
				//day
				var _, _, lDay = l.initTime.Date()
				if lDay < day {
					bakPath = fmt.Sprintf(bakFmt, name, l.initTime.Format(backupTypeName[0:len(backupTypeName)-9]))
					l.initTime = time
				}
			case "M":
				//month
				var _, lMonth, _ = l.initTime.Date()
				if lMonth < month {
					bakPath = fmt.Sprintf(bakFmt, name, l.initTime.Format(backupTypeName[0:len(backupTypeName)-12]))
					l.initTime = time
				}

			case "Y":
				//year
				var lYear, _, _ = l.initTime.Date()
				if lYear < year {
					bakPath = fmt.Sprintf(bakFmt, name, l.initTime.Format(backupTypeName[0:len(backupTypeName)-15]))
					l.initTime = time
				}
			default:
				//pass
			}

			if len(bakPath) > 0 {
				var newFile, err = backup(name, string(bakPath))
				if err != nil {
					return i + 1, err
				}

				if newFile != nil {
					var errC = f.Close()
					if errC != nil {
						return i + 1, errC
					}

					l.outputs[i] = newFile
					f = newFile
				}
			}

			finalBuf = append(finalBuf, *b...)
		}

		var _, err = f.Write(finalBuf)
		if err != nil {
			return i + 1, err
		}
	}

	return 0, nil
}

func backup(oldPath string, newPath string) (*os.File, error) {
	if !checkPathExists(newPath) {
		var errRn = os.Rename(oldPath, newPath)
		if errRn != nil {
			return nil, errRn
		}

		var newFile, errNf = createFile(oldPath)
		if errNf != nil {
			return nil, errNf
		}

		return newFile, nil
	}

	return nil, nil
}

func checkPathExists(path string) bool {
	var _, err = os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsExist(err) {
		return true
	}

	return false
}

func createFile(path string) (*os.File, error) {
	var file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	return file, err
}

// 初始化设置

// New 日志实体初始化
func New() *Logger {
	var logger = &Logger{
		outputs:    []*os.File{os.Stdout},
		level:      LevelDebug,
		timeFormat: "2006-01-02 15:04:05.000",
		backupType: "D",
		maxSize:    1024 * 1024 * 1024 * 50,
		callDepth:  2,
		initTime:   time.Now(),
	}

	return logger
}

// SetOutputs
// path: 多个文件路径, 用逗号“,”隔开
func (l *Logger) SetOutputs(paths string) *Logger {
	var fs []*os.File

	var ps = strings.Split(paths, ",")
	for _, path := range ps {
		switch path {
		case "":
		case "stdout":
			fs = append(fs, os.Stdout)
		case "stderr":
			fs = append(fs, os.Stderr)
		default:
			var f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}

			fs = append(fs, f)
		}
	}

	logger.outputs = fs
	return l
}

// SetLevel 设置日志级别
// level: 请传入常量参数 example: LevelDebug
func (l *Logger) SetLevel(level int) *Logger {
	if level < LevelDebug || level > LevelFatal {
		panic("unsupport level")
	}

	l.level = level
	return l
}

// SetTimeFormat 设置时间显示格式
// format : example : 2006-01-02 15:04:05.000
func (l *Logger) SetTimeFormat(format string) *Logger {
	l.timeFormat = format
	return l
}

// SetBackupType 设置备份类型
// tp:
//	size: 大小 (单位byte)
//	s: 每秒
//	m: 每分
//	h: 每小时
//	D: 每天
//	M: 每月
//	Y: 每年
func (l *Logger) SetBackupType(tp string) *Logger {
	if !(tp == "size" ||
		tp == "s" ||
		tp == "m" ||
		tp == "h" ||
		tp == "D" ||
		tp == "M" ||
		tp == "Y") {
		panic("unsupport backup type")
	}
	logger.backupType = tp
	return l
}

// SetMaxSize 如果backup type为size, 需要设置文件最大限制
// size: 文件最大大小 (单位size), example: 1024*1024*1024*50 = 50MB
func (l *Logger) SetMaxSize(size int64) *Logger {
	if size <= 0 {
		panic("invalid file size")
	}

	l.maxSize = size
	return l
}

// Debug log
func (l *Logger) Debug(v ...interface{}) {
	var s = fmt.Sprintln(v...)
	l.output(PrefixDebug, s, LevelDebug, TerminalColorGreen)
}

// Debugf log
func (l *Logger) Debugf(format string, v ...interface{}) {
	var s = fmt.Sprintln(fmt.Sprintf(format, v...))
	l.output(PrefixDebug, s, LevelDebug, TerminalColorGreen)
}

// Info log
func (l *Logger) Info(v ...interface{}) {
	var s = fmt.Sprintln(v...)
	l.output(PrefixInfo, s, LevelInfo, TerminalColorWhite)
}

// Infof log
func (l *Logger) Infof(format string, v ...interface{}) {
	var s = fmt.Sprintln(fmt.Sprintf(format, v...))
	l.output(PrefixInfo, s, LevelInfo, TerminalColorWhite)
}

// Warn log
func (l *Logger) Warn(v ...interface{}) {
	var s = fmt.Sprintln(v...)
	l.output(PrefixWarn, s, LevelWarn, TerminalColorYellow)
}

// Warnf log
func (l *Logger) Warnf(format string, v ...interface{}) {
	var s = fmt.Sprintln(fmt.Sprintf(format, v...))
	l.output(PrefixWarn, s, LevelWarn, TerminalColorYellow)
}

// Error log
func (l *Logger) Error(v ...interface{}) {
	var s = fmt.Sprintln(v...)
	l.output(PrefixError, s, LevelError, TerminalColorRed)
}

// Errorf log
func (l *Logger) Errorf(format string, v ...interface{}) {
	var s = fmt.Sprintln(fmt.Sprintf(format, v...))
	l.output(PrefixError, s, LevelError, TerminalColorRed)
}

// Fatal log with panic
func (l *Logger) Fatal(v ...interface{}) {
	var s = fmt.Sprintln(v...)
	l.output(PrefixFatal, s, LevelFatal, TerminalColorMagenta)

	panic(s)
	//os.Exit(1)
}

// Fatalf log with panic
func (l *Logger) Fatalf(format string, v ...interface{}) {
	var s = fmt.Sprintln(fmt.Sprintf(format, v...))
	l.output(PrefixFatal, s, LevelFatal, TerminalColorMagenta)

	panic(s)
}

var logger *Logger

// Init 初始化
func Init() *Logger {
	logger = New()
	return logger
}

// Debug log
func Debug(v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(v...)
		logger.output(PrefixDebug, s, LevelDebug, TerminalColorGreen)
	}
}

// Debugf log
func Debugf(format string, v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(fmt.Sprintf(format, v...))
		logger.output(PrefixDebug, s, LevelDebug, TerminalColorGreen)
	}
}

// Info log
func Info(v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(v...)
		logger.output(PrefixInfo, s, LevelInfo, TerminalColorWhite)
	}
}

// Infof log
func Infof(format string, v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(fmt.Sprintf(format, v...))
		logger.output(PrefixInfo, s, LevelInfo, TerminalColorWhite)
	}
}

// Warn log
func Warn(v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(v...)
		logger.output(PrefixWarn, s, LevelWarn, TerminalColorYellow)
	}
}

// Warnf log
func Warnf(format string, v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(fmt.Sprintf(format, v...))
		logger.output(PrefixWarn, s, LevelWarn, TerminalColorYellow)
	}
}

// Error log
func Error(v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(v...)
		logger.output(PrefixError, s, LevelError, TerminalColorRed)
	}
}

// Errorf log
func Errorf(format string, v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(fmt.Sprintf(format, v...))
		logger.output(PrefixError, s, LevelError, TerminalColorRed)
	}
}

// Fatal log with panic
func Fatal(v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(v...)
		logger.output(PrefixFatal, s, LevelFatal, TerminalColorMagenta)

		panic(s)
		//os.Exit(1)
	}

}

// Fatalf log with panic
func Fatalf(format string, v ...interface{}) {
	if logger != nil {
		var s = fmt.Sprintln(fmt.Sprintf(format, v...))
		logger.output(PrefixFatal, s, LevelFatal, TerminalColorMagenta)

		panic(s)
	}
}

// StackTrace print
func StackTrace(message string, skip int, maxLen int) {
	var pcs = make([]uintptr, maxLen)
	var n = runtime.Callers(skip, pcs)

	var info []string
	for i, pc := range pcs[:n] {
		var funcPC = runtime.FuncForPC(pc)
		var file, line = funcPC.FileLine(pc)

		info = append(info, fmt.Sprintf("(%d)[%s]%s:%d", i, funcPC.Name(), file, line))
	}

	Debug(fmt.Sprintf("%s\nStackTrace:\n%s\n", message, strings.Join(info, "\n")))
}
