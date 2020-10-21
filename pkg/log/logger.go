package log

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

var logger = &Logger{}

type Logger struct {
	basicLogger *log.Logger
}

func GetFromAddress(host string, ip string) string {
	if host == "" || ip == "" {
		return "from=<unknown address>"
	}

	return fmt.Sprintf("from=<%s(%s)>", host, ip)
}

func SetFormat(packageName string, debug bool) {
	logger = &Logger{
		basicLogger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
	logger.setFormat(packageName, debug)
}

func Get() *log.Logger {
	return logger.basicLogger
}

func Debug(format string) {
	logger.debug(format)
}

func Info(format string) {
	logger.info(format)
}

func Debugf(format string, v ...interface{}) {
	logger.debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	logger.infof(format, v...)
}

func Errorf(format string, v ...interface{}) {
	logger.errorf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	logger.fatalf(format, v...)
	os.Exit(1)
}

func (l *Logger) setFormat(packageName string, debug bool) {
	if !debug {
		l.basicLogger.SetOutput(
			&logutils.LevelFilter{
				Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR", "FATAL"},
				MinLevel: logutils.LogLevel("INFO"),
				Writer:   os.Stderr,
			},
		)
	}
}

func (l *Logger) debug(format string) {
	l.basicLogger.Print("[DEBUG] " + format)
}

func (l *Logger) info(format string) {
	l.basicLogger.Print("[INFO] " + format)
}

func (l *Logger) debugf(format string, v ...interface{}) {
	l.basicLogger.Printf("[DEBUG] "+format, v...)
}

func (l *Logger) infof(format string, v ...interface{}) {
	l.basicLogger.Printf("[INFO] "+format, v...)
}

func (l *Logger) errorf(format string, v ...interface{}) {
	l.basicLogger.Printf("[ERROR] "+format, v...)
}

func (l *Logger) fatalf(format string, v ...interface{}) {
	l.basicLogger.Printf("[FATAL] "+format, v...)
	os.Exit(1)
}
