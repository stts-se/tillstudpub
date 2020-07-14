// Package logger is a simple logger API to make global logger configuration easier in future versions
package logger

import (
	"fmt"
	"os"

	l "log"
)

const (
	debug   = "debug"
	error   = "error"
	info    = "info"
	warning = "warning"
)

// Debug logs a message with 'debug' level
func Debug(v ...interface{}) {
	log(debug, fmt.Sprint(v...))
}

// Debugf logs a message with 'debug' level
func Debugf(format string, v ...interface{}) {
	log(debug, fmt.Sprintf(format, v...))
}

// Info logs a message with 'info' level
func Info(v ...interface{}) {
	log(info, fmt.Sprint(v...))
}

// Infof logs a message with 'info' level
func Infof(format string, v ...interface{}) {
	log(info, fmt.Sprintf(format, v...))
}

// Error logs a message with 'error' level
func Error(v ...interface{}) {
	log(error, fmt.Sprint(v...))
}

// Errorf logs a message with 'error' level
func Errorf(format string, v ...interface{}) {
	log(error, fmt.Sprintf(format, v...))
}

// Warning logs a messages with 'warning' level
func Warning(v ...interface{}) {
	log(warning, fmt.Sprint(v...))
}

// Warningf logs a message with 'warning' level
func Warningf(format string, v ...interface{}) {
	log(warning, fmt.Sprintf(format, v...))
}

// Fatal is equivalent to Error() followed by a call to os.Exit(1)
func Fatal(v ...interface{}) {
	log(error, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Errorf() followed by a call to os.Exit(1)
func Fatalf(format string, v ...interface{}) {
	log(error, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func log(level, message string) {
	l.Printf("%s: %s", level, message)
}
