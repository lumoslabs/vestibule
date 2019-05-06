package log

import "fmt"

// Logger is a simple interface that handles Info and Debug logging
type Logger interface {
	Info(string)
	Infof(string, ...interface{})
	Debug(string)
	Debugf(string, ...interface{})
	IsDebug() bool
}

// GetLogger returns the package logger
func GetLogger() Logger { return logger }

// SetLogger sets the package logger
func SetLogger(l Logger) { logger = l }

// Info writes info level messages using the package logger
func Info(msg string) { logger.Info(msg) }

// Infof writes formatted info level messages with the package logger
func Infof(fmt string, inf ...interface{}) { logger.Infof(fmt, inf...) }

// Debug writes debug level messages using the package logger
func Debug(msg string) { logger.Debug(msg) }

// Debugf writes formatted debug level messages using the package logger
func Debugf(fmt string, inf ...interface{}) { logger.Debugf(fmt, inf...) }

// IsDebug returns true if the package logger is at Debug level
func IsDebug() bool { return logger.IsDebug() }

var logger Logger = NewNilLogger()

type nilLogger bool

func (nl *nilLogger) Info(s string)                     {}
func (nl *nilLogger) Infof(f string, o ...interface{})  {}
func (nl *nilLogger) Debug(s string)                    {}
func (nl *nilLogger) Debugf(f string, o ...interface{}) {}
func (nl *nilLogger) IsDebug() bool                     { return false }

func NewNilLogger() Logger { return new(nilLogger) }

type debugLogger bool

func (dl *debugLogger) Info(s string)                     { fmt.Println("[inf] " + s) }
func (dl *debugLogger) Infof(f string, o ...interface{})  { fmt.Println(fmt.Sprintf("[inf] "+f, o...)) }
func (dl *debugLogger) Debug(s string)                    { fmt.Println("[dbg] " + s) }
func (dl *debugLogger) Debugf(f string, o ...interface{}) { fmt.Println(fmt.Sprintf("[dbg] "+f, o...)) }
func (dl *debugLogger) IsDebug() bool                     { return true }

func NewDebugLogger() Logger { return new(debugLogger) }
