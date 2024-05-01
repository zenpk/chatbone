package util

import (
	"fmt"
	"log"
	"os"
	"time"
)

type ILogger interface {
	Printf(format string, args ...interface{})
	Println(any ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(any ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(any ...interface{})
	Close() error
}

type Logger struct {
	*log.Logger
	logFile *os.File
}

func NewLogger(conf *Configuration) (*Logger, error) {
	l := new(Logger)
	l.Logger = new(log.Logger)
	logFile, err := os.OpenFile(conf.LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return l, nil
	}
	l.logFile = logFile
	l.Logger.SetOutput(logFile)
	// remove timestamp
	l.Logger.SetFlags(0)
	return l, nil
}

func (l *Logger) Printf(format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)
	l.printWithTime(content)
}

func (l *Logger) Println(any ...interface{}) {
	content := fmt.Sprintln(any...)
	l.printWithTime(content)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	content := fmt.Sprintf("WARN : "+format, args...)
	l.printWithTime(content)
}

func (l *Logger) Warnln(any ...interface{}) {
	content := fmt.Sprintln(append([]interface{}{"WARN :"}, any...)...)
	l.printWithTime(content)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	content := fmt.Sprintf("ERROR: "+format, args...)
	l.printWithTime(content)
}

func (l *Logger) Errorln(any ...interface{}) {
	content := fmt.Sprintln(append([]interface{}{"ERROR:"}, any...)...)
	l.printWithTime(content)
}

func (l *Logger) Close() error {
	return l.logFile.Close()
}

func (l *Logger) printWithTime(content string) {
	const Format = "2006-01-02 15:04:05"
	l.Logger.Printf("[%s] %s", time.Now().UTC().Format(Format), content)
}
