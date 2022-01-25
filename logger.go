package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

// These flags define which text to prefix to each log entry generated by the Logger.
const (
	// Bits or'ed together to control what's printed. There is no control over the
	// order they appear (the order listed here) or the format they present (as
	// described in the comments).  A colon appears after these items:
	//	[log] 2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
	Ldate         = log.Ldate                                                         // the date: 2009/01/23
	Ltime         = log.Ltime                                                         // the time: 01:23:23
	Lmicroseconds = log.Lmicroseconds                                                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     = log.Llongfile                                                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    = log.Lshortfile                                                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC          = log.LUTC                                                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lpackage      = log.LUTC << 1                                                     // package name: [log]
	Lfunction     = Lpackage << 1                                                     // function name: <Print>
	LstdFlags     = Ldate | Ltime | Lmicroseconds | Lshortfile | Lpackage | Lfunction // initial values for the standard logger
	Ldevelop      = Ldate | Ltime | Lmicroseconds | Llongfile | Lpackage | Lfunction
	Lproduction   = Ldate | Ltime | Lmicroseconds | Lpackage | Lfunction
)

//日志级别
const (
	LEVEL_DISABLE = 0 //关闭日志功能
	LEVEL_DEBUG   = 1 << iota
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
	LEVEL_CRITICAL
	LEVEL_PANIC
)

const (
	LEVEL_FATAL = "[FATAL]"
	LEVEL_ALL   = LEVEL_DEBUG | LEVEL_INFO | LEVEL_WARN | LEVEL_ERROR | LEVEL_CRITICAL | LEVEL_PANIC

	//默认日志级别为
	LEVEL_DEFAULT = LEVEL_ALL
)

var (
	LevelText = map[string]int{
		"disable":  LEVEL_DISABLE,
		"debug":    LEVEL_DEBUG,
		"info":     LEVEL_INFO,
		"warn":     LEVEL_WARN,
		"error":    LEVEL_ERROR,
		"critical": LEVEL_CRITICAL,
		"panic":    LEVEL_PANIC,
		"all":      LEVEL_ALL,
	}

	logPrefixs = map[int]string{
		LEVEL_DEBUG:    "[DEBUG]",
		LEVEL_INFO:     "[INFO]",
		LEVEL_WARN:     "[WARN]",
		LEVEL_ERROR:    "[ERROR]",
		LEVEL_CRITICAL: "[CRITICAL]",
		LEVEL_PANIC:    "[PANIC]",
	}
)

type Logger struct {
	*log.Logger
	mu    sync.Mutex
	level int
}

func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{
		Logger: log.New(out, prefix, flag),
		level:  LEVEL_DEFAULT,
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	prefix := l.Prefix()
	flags := l.Flags()
	levels := l.Levels()
	l.mu.Unlock()
	*l = Logger{
		Logger: log.New(w, prefix, flags),
		level:  levels,
	}
}

func ParseLevel(level string) int {
	if level == "" {
		return LEVEL_DEFAULT
	}
	var lv int
	lvs := strings.Split(level, ",")
	for _, v := range lvs {
		if l, lok := LevelText[strings.ToLower(strings.TrimSpace(v))]; lok {
			lv |= l
		}
	}

	return lv
}

func (l *Logger) SetLevel(level interface{}) {
	switch v := level.(type) {
	case int:
		l.mu.Lock()
		defer l.mu.Unlock()
		l.level = v
	case string:
		lv := ParseLevel(v)
		l.mu.Lock()
		defer l.mu.Unlock()
		l.level = lv
	}
}

func (l *Logger) Levels() int {
	return l.level
}

func (l *Logger) Output(calldepth int, s string) error {
	if Lpackage&l.Flags() != 0 {
		if pc, _, _, ok := runtime.Caller(calldepth); ok {
			pkg_func := runtime.FuncForPC(pc).Name()
			if pos := strings.LastIndex(pkg_func, "."); pos != -1 {
				if pos1 := strings.LastIndex(pkg_func, ".("); pos1 != -1 {
					pos = pos1
				}

				pkg, fc := pkg_func[:pos], pkg_func[pos+1:]
				if Lfunction&l.Flags() != 0 {
					l.SetPrefix(l.Prefix() + "[" + pkg + "] <" + fc + "> ")
				} else {
					l.SetPrefix(l.Prefix() + "[" + pkg + "] ")
				}
			} else {
				if Lfunction&l.Flags() != 0 {
					l.SetPrefix(l.Prefix() + "<" + pkg_func + "> ")
				}
			}
		}
	}

	//再加1层，是因为l.loger调用
	calldepth = calldepth + 1
	return l.Logger.Output(calldepth, s)
}

func (l *Logger) print(prefix string, v ...interface{}) {
	//使用内部函数供可导出函数使用,所有又包了2层
	localCalldepth := 2 + 2
	s := fmt.Sprint(v...)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.SetPrefix(prefix)
	l.Output(localCalldepth, s)
}

func (l *Logger) println(prefix string, v ...interface{}) {
	//使用内部函数供可导出函数使用,所有又包了2层
	localCalldepth := 2 + 2
	s := fmt.Sprintln(v...)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.SetPrefix(prefix)
	l.Output(localCalldepth, s)
}

func (l *Logger) printf(prefix string, format string, v ...interface{}) {
	//使用内部函数供可导出函数使用,所有又包了2层
	localCalldepth := 2 + 2
	s := fmt.Sprintf(format, v...)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.SetPrefix(prefix)
	l.Output(localCalldepth, s)
}

func (l *Logger) Print(v ...interface{}) {
	l.print("", v...)
}

func (l *Logger) Println(v ...interface{}) {
	l.println("", v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.printf("", format, v...)
}

func (l *Logger) Panic(v ...interface{}) {
	if LEVEL_PANIC&l.level == 0 {
		return
	}

	l.print(logPrefixs[LEVEL_PANIC]+" ", v...)
	panic(fmt.Sprint(v...))
}

func (l *Logger) Panicln(v ...interface{}) {
	if LEVEL_PANIC&l.level == 0 {
		return
	}

	l.println(logPrefixs[LEVEL_PANIC]+" ", v...)
	panic(fmt.Sprintln(v...))
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	if LEVEL_PANIC&l.level == 0 {
		return
	}

	l.printf(logPrefixs[LEVEL_PANIC]+" ", format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.print(LEVEL_FATAL+" ", v...)
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.println(LEVEL_FATAL+" ", v...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.printf(LEVEL_FATAL+" ", format, v...)
	os.Exit(1)
}
