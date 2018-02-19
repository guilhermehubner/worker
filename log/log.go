package log

import "fmt"

var logger LoggerInterface

type LoggerInterface interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type defaultLogger struct{}

func init() {
	logger = defaultLogger{}
}

func (l defaultLogger) Info(args ...interface{}) {
	fmt.Printf("\x1b[1;34m[INFO] \x1b[0m%s\n", fmt.Sprint(args...))
}

func (l defaultLogger) Error(args ...interface{}) {
	fmt.Printf("\x1b[1;31m[ERROR] \x1b[0m%s\n", fmt.Sprint(args...))
}

// Get gets the logger
func Get() LoggerInterface {
	return logger
}

// Set sets a custom logger
func Set(l LoggerInterface) {
	logger = l
}
