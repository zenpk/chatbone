package util

import (
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

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.printTime()
	l.Printf("WARN : "+format, args...)
}

func (l *Logger) Warnln(any ...interface{}) {
	l.printTime()
	l.Print(append([]interface{}{"WARN :"}, any...)...)
	l.Print("\n")
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.printTime()
	l.Printf("ERROR: "+format, args...)
}

func (l *Logger) Errorln(any ...interface{}) {
	l.printTime()
	l.Print(append([]interface{}{"ERROR:"}, any...)...)
	l.Print("\n")
}

func (l *Logger) Close() error {
	return l.logFile.Close()
}

func (l *Logger) printTime() {
	const Format = "2006-01-02 15:04:05"
	l.Printf("[%s] ", time.Now().UTC().Format(Format))
}
