package logging

import (
	"fmt"
	"log"
	"os"
	"io"
	"time"
	"path/filepath"
	"strings"
	"text/template"
	"bytes"
	"sync"
	"errors"
)

const (
	FATAL = iota
	ERROR
	WARN
	INFO
	DEBUG
)

//wrapping log flags from go log lib
const (
	Ldate = log.Ldate
	Ltime = log.Ltime
	Lshortfile = log.Lshortfile
	Llongfile = log.Llongfile
	Lmicroseconds        = log.Lmicroseconds
	LUTC = log.LUTC
	LstdFlags = log.LstdFlags
)

type Logger struct {
	*log.Logger
	
	logPath    string
	logFilePrefix    string
	logFileType    string
	logLevel   int
	logFlags	int
	current  io.Writer
	//rotate
	lastRotateTime time.Time
	rotationInterval time.Duration
	alignTime bool
	mutex	sync.Mutex
	parent  *Logger
}

//The variables that allowed to be used as template variable in logger path parameter
type LogPathVariables struct {
	Date string
	//TODO: to be added
	Hostname string
	//Just for testing & demo purpose
	CorpName string
}

func evaluateLogPath(path string, flagUTC int) (evaluatedPath string) {
	now := time.Now()
	if flagUTC != 0 {
		now = now.UTC()
	}
	year, month, day := now.Date()
	lpv := &LogPathVariables{Date:fmt.Sprintf("%04d%02d%02d", year, month, day), CorpName: "supersearch"}
	if hostname, err := os.Hostname(); err == nil {
		lpv.Hostname = hostname
	}
	var buf bytes.Buffer
	if tmpl, err := template.New("path").Parse(path); err == nil {
		tmpl.Execute(&buf, lpv)
		evaluatedPath = buf.String()
	} else {
		evaluatedPath = path
	}
	return
}

func EnsureDir(dir string) error {
	if err := os.MkdirAll(dir, 0777); err != nil {
		if !os.IsExist(err) {
			log.Println("Failed to create path " + dir)
			return err
		}
	}
	return nil
}

func ensureLogDir(dir string) error {
	return EnsureDir(dir)
}

func AlignTime(t time.Time, d time.Duration) (time.Time) {
	if t.Location() == time.UTC {
		return t.Truncate(d)
	}

	yy, mm, dd := t.Date()
	switch d {
	case ROTATE_DAILY:
		t = time.Date(yy, mm, dd, 0, 0, 0, 0, t.Location())
	case ROTATE_HOURLY:
		t = time.Date(yy, mm, dd, t.Hour(), 0, 0, 0, t.Location())
	}
	return t
}

func (l *Logger) getLogSuffixNanoSecond() string {
	now := time.Now()
	if l.logFlags&LUTC != 0 {
		now = now.UTC()
	}
	if l.alignTime {
		now = AlignTime(now, l.rotationInterval)
	}
	var fileType = "log"
	if len(l.logFileType) != 0 {
		fileType = l.logFileType
	}
	year, month, day := now.Date()
	return fmt.Sprintf("-%04d%02d%02d%02d%02d%02d_%x.%s", year, month, day, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), fileType)
}

// Init and return a new file based logger with given path and file name
func NewLogger(path, file, fileType string, level, flags int, extraConfig ...interface{}) (logger *Logger, e error) {
	if len(extraConfig) > 1 {
		e = errors.New("NewLogger only supports one extral configuration: alignMargin (time.Duration)")
		return
	}
	var alignTime = false
	var alignMargin = time.Duration(0)
	if len(extraConfig) == 1 {
		var ok bool
		if alignMargin, ok = extraConfig[0].(time.Duration); !ok {
			e = errors.New("extral configuration shoud be time.Duration type")
			return
		}
		if alignMargin > 0 {
			alignTime = true
			flags |= LUTC
		}
	}
	dir := evaluateLogPath(path, flags|LUTC)
	if e = ensureLogDir(dir); e != nil {
		return
	}
	file = strings.Replace(file, string(filepath.Separator), "-", -1)
	logger = &Logger{logPath:path, logFilePrefix:file, logFileType: fileType, logLevel:level, logFlags:flags}
	logger.alignTime = alignTime
	logger.rotationInterval = alignMargin
	defer func(){
		logger.alignTime = false
		logger.rotationInterval = time.Duration(0)
	}()
	logSuffix := logger.getLogSuffixNanoSecond()
	logFile, err := os.OpenFile(filepath.Join(dir, logger.logFilePrefix+logSuffix), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		e = err
		return
	}
	logger.current = logFile
	logger.Logger = log.New(logFile, "", flags)
	return
}

func DecorateDefaultLogger(categories ...string) (logger *Logger) {
	
	var categoriesStr string
	
	if len(categories) > 0  {
		var categoryBuffer bytes.Buffer
		for _, category := range categories {
			categoryBuffer.WriteString(fmt.Sprintf("[%s]", category))
		}
		categoriesStr = categoryBuffer.String()
	}
	lg := log.New(std.current, fmt.Sprintf("%s ", categoriesStr), std.logFlags)
	
	logger = &Logger{logPath:std.logPath,
		logFilePrefix:std.logFilePrefix,
		logFileType:std.logFileType,
		logFlags:std.logFlags,
		logLevel:std.logLevel,
		Logger:lg}
	return
}

func NewLogFile(filename string) (logger *Logger, e error) {
	if e = ensureLogDir(filepath.Dir(filename)); e != nil {
		return
	}
	logger = &Logger{}
	logFile, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		e = err
		return
	}

	logger.current = logFile
	logger.Logger = log.New(logFile, "", 0)
	return
}

//Rotate log file
func (l *Logger) Rotate() (e error) {
	dir := evaluateLogPath(l.logPath, l.logFlags&LUTC)
	if e = ensureLogDir(dir); e != nil {
		return
	}
	logSuffix := l.getLogSuffixNanoSecond()

	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.rotateLogFile(filepath.Join(dir, l.logFilePrefix+logSuffix))
}

func (l *Logger) rotateLogFile( filename string ) (e error){
	var logFile *os.File
	if logFile, e = os.OpenFile( filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); e == nil {
		l.Logger = log.New(logFile, "", l.logFlags)
		currentFile := l.current.(*os.File)
		currentFile.Sync()
		currentFile.Close()
		l.current = logFile
		l.lastRotateTime = time.Now()
		if l.logFlags&LUTC != 0 {
			l.lastRotateTime = l.lastRotateTime.UTC()
		}
		if l.alignTime {
			l.lastRotateTime = AlignTime(l.lastRotateTime, l.rotationInterval)
		}
	} else {
		l.Error("error rotating log file", e.Error())
		return
	}
	return
}

func (l *Logger) RotationInterval(d time.Duration) {
	l.rotationInterval = d
}

//Close function is provided for transactional log that should close file handle immediately after logging done.
func (l *Logger) Close() (e error) {
	l.Logger = nil
	if l.current != nil {
		currentFile := l.current.(*os.File)
		currentFile.Sync()
		e = currentFile.Close()
	}
	return
}

//Return current log file name 
func (l *Logger) GetLogFile() string {
	currentFile := l.current.(*os.File)
	return currentFile.Name()
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= DEBUG {
		l.Output(2, fmt.Sprintf("[DEBUG] "+format+"\n", v...))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= INFO {
		l.Output(2, fmt.Sprintf("[INFO] "+format+"\n", v...))
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= WARN {
		l.Output(2, fmt.Sprintf("[WARNING] "+format+"\n", v...))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= ERROR {
		l.Output(2, fmt.Sprintf("[ERROR] "+format+"\n", v...))
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= FATAL {
		l.Output(2, fmt.Sprintf("[FATAL] "+format+"\n", v...))
		os.Exit(1)
	}
}

// DebugLv print the log if loglevel>= DEBUG with the callstack depth equals to stacklv.
func (l *Logger) DebugLv(stacklv int,format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= DEBUG {
		l.Output(stacklv, fmt.Sprintf("[DEBUG] "+format+"\n", v...))
	}
}

// InfoLv print the log if loglevel>= INFO with the callstack depth equals to stacklv.
func (l *Logger) InfoLv(stacklv int,format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= INFO {
		l.Output(stacklv, fmt.Sprintf("[INFO] "+format+"\n", v...))
	}
}

// WarnLv print the log if loglevel>= WARN with the callstack depth equals to stacklv.
func (l *Logger) WarnLv(stacklv int,format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= WARN {
		l.Output(stacklv, fmt.Sprintf("[WARNING] "+format+"\n", v...))
	}
}

// ErrorLv print the log if loglevel>= ERROR with the callstack depth equals to stacklv.
func (l *Logger) ErrorLv(stacklv int,format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= ERROR {
		l.Output(stacklv, fmt.Sprintf("[ERROR] "+format+"\n", v...))
	}
}

// FatalLv print the log if loglevel>= FATAL with the callstack depth equals to stacklv.
func (l *Logger) FatalLv(stacklv int,format string, v ...interface{}) {
	if l.Logger != nil && l.logLevel >= FATAL {
		l.Output(stacklv, fmt.Sprintf("[FATAL] "+format+"\n", v...))
		os.Exit(1)
	}
}
//Output formatted raw string to log file, for transacton logs
func (l *Logger) RawF(format string, v ...interface{}) (n int, err error) {
	if l.current != nil {
		currentFile := l.current.(*os.File)
		return currentFile.WriteString(fmt.Sprintf(format,v))
	}
	return
}

//Output raw string to log file, for transacton logs
func (l *Logger) Raw(content string) (n int, err error){
	if l.current != nil {
		currentFile := l.current.(*os.File)
		return currentFile.WriteString(content)
	}
	return
}
