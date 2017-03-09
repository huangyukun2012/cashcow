package logging

import (
	"time"
	"log"
	"os"
	_"io"
	"fmt"
	"bytes"
)

const (
	ROTATE_HOURLY = time.Duration(3600) * time.Second
	ROTATE_DAILY = time.Duration(86400) * time.Second
)

type LogService struct {
	exit chan int
	loggers map[string]*Logger
	ticker *time.Ticker
	useUTC bool
	alignTime bool
}

var logStatusFuncMap = map[int]string{
	FATAL: "Fatal",
	ERROR: "Error",
	WARN:  "Warn",
	INFO:  "Info",
	DEBUG: "Debug",
}

//Create new instance for log service
func NewLogServcie() *LogService {
	return &LogService{exit:make(chan int), loggers:make(map[string]*Logger)}
}

var LOG_LEVEL = INFO
//Global logger for service
var std = &Logger{Logger:log.New(os.Stderr,"",log.LstdFlags), logLevel:LOG_LEVEL}

//Config LogService whether to use UTC time
//This function should be called just after NewLogServcie()
func (p *LogService) UseUTC() (*LogService) {
	p.useUTC = true
	return p
}

//Config LogService whether to keep log rotation time align to integral time point
//Current log service only supports alignment of time for UTC. If this function is called, it will force log service to use UTC. 
//This function should be called just after NewLogServcie()
func (p *LogService) RotateAlignToTime() (*LogService) {
	p.UseUTC()
	p.alignTime = true
	return p
}

//Config default to write to file and to auto-rotate
func (p *LogService) ConfigDefaultLogger(path, namePrefix string, logLevel int, rotationInterval time.Duration) (err error) {
	var logger *Logger
	flags := LstdFlags
	if logLevel >= DEBUG {
		flags |= Lshortfile
	}
	if p.useUTC {
		flags |= LUTC
	}
	var alignMargin = time.Duration(0)
	if p.alignTime {
		alignMargin = rotationInterval
	}
	if logger, err = NewLogger(path, namePrefix, "log", logLevel, flags, alignMargin); err == nil {
		std = logger
		p.RegisterLogger("default", std, rotationInterval)
	}
	LOG_LEVEL = logLevel
	return
}

//Start log service, returns immediately. 
//Service loop will be run in a goroutine, handling logger rotation
func (p *LogService) Serve() {
	p.ticker = time.NewTicker(time.Duration(1) * time.Minute)
	go func(p *LogService) {
		for {
			select {
			case <-p.ticker.C:
				p.onTimer()
			case <-p.exit:
				p.ticker.Stop()
				break
			}
		}
	}(p)
}

func (p *LogService) onTimer() {
	now := time.Now()
	if p.useUTC {
		now = now.UTC()
	}
	for _, logger := range p.loggers {
		if logger.rotationInterval > time.Duration(0) && logger.lastRotateTime.Add(logger.rotationInterval).Before(now) {
			currentFile, _ := logger.current.(*os.File)
			logger.Info("Rotating log file: %s", currentFile.Name())
			logger.Rotate()
		}
	}
}

//Register Logger to LogServcie so it can be auto-rotated
//to disable the log rotation, use 0 as rotationInterval
func (p *LogService) RegisterLogger(name string, logger *Logger, rotationInterval time.Duration)  {
	if logger != nil {
		if p.useUTC {
			logger.logFlags = logger.logFlags|LUTC
			logger.Logger.SetFlags(logger.Logger.Flags()|LUTC)
		}
		if p.alignTime && rotationInterval > 0 {
			logger.alignTime = true
		}
		if rotationInterval > 0 {
			logger.lastRotateTime = time.Now()
			if logger.logFlags&LUTC != 0 {
				logger.lastRotateTime = logger.lastRotateTime.UTC()
			}
			if logger.alignTime {
				logger.lastRotateTime = AlignTime(logger.lastRotateTime, rotationInterval)
			}
			logger.RotationInterval(rotationInterval)
		}
		p.loggers[name] = logger
	}
}

//Get logger by name, return nil when no logger with such name registered
func (p *LogService) GetLogger(name string) (logger *Logger) {
	logger, _ = p.loggers[name]
	return
}

func (p *LogService) Stop() {
	for _, logger := range p.loggers {
		logger.Close()
	}
	p.exit<-1
}

//Wrapping package level function for global logger output.
//We didn't wrap log.Panic because we DON'T PANIC.
func Debug(format string, v ...interface{}) {
	if std.Logger != nil && LOG_LEVEL >= DEBUG {
		Printf("[DEBUG] "+format+"\n", v...)
	}
}

func Info(format string, v ...interface{}) {
	if std.Logger != nil && LOG_LEVEL >= INFO {
		Printf("[INFO] "+format+"\n", v...)
	}
}

func Warn(format string, v ...interface{}) {
	if std.Logger != nil && LOG_LEVEL >= WARN {
		Printf("[WARNING] "+format+"\n", v...)
	}
}

func Error(format string, v ...interface{}) {
	if std.Logger != nil && LOG_LEVEL >= ERROR {
		Printf("[ERROR] "+format+"\n", v...)
	}
}

func Panicf(format string, v ...interface{}) {
	std.Panicf("[Panic] "+format+"\n", v...)
}

func Panic(v ...interface{}) {
	std.Panic(v...)
}

func Fatal(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf("[FATAL] "+format+"\n", v...))
	os.Exit(1)
}

func Log(statusCode int, categories []string, format string, v ...interface{})  {
	
	//Don't do anything
	if LOG_LEVEL < statusCode {
		return 
	}
	
	status, _ := logStatusFuncMap[statusCode]; 
	var categoriesStr = ""
	
	if len(categories) > 0 {
			var categoryBuffer bytes.Buffer
			for _, category := range categories {
				categoryBuffer.WriteString(fmt.Sprintf("[%s]", category))
			}
			categoriesStr = categoryBuffer.String()
	}
	std.Output(3, categoriesStr+"["+status+"] "+ fmt.Sprintf(format, v...))
}

//Output formatted raw string to log file, for transacton logs
func RawF(format string, v ...interface{}) {
	std.RawF(format, v)
}

//Output raw string to log file, for transacton logs
func Raw(content string) {
	std.Raw(content)
}

//for ease of migration
func Printf(format string, v ...interface{}) {
	if std.Logger != nil {
		std.Output(3, fmt.Sprintf(format, v...))
	}
}

func Println(v ...interface{}) {
	if std.Logger != nil {
		std.Println(v...)
	}
}
