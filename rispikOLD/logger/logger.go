// Package logger is a simple logger API to make global logger configuration easier in future versions
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/stts-se/tillstudpub/rispik/protocol"
)

var wsListeners = map[*websocket.Conn]bool{}
var wsMux sync.RWMutex

func AddWSListener(conn *websocket.Conn) {
	wsMux.Lock()
	defer wsMux.Unlock()
	wsListeners[conn] = true
}

func removeWSListener(conn *websocket.Conn) {
	wsMux.Lock()
	defer wsMux.Unlock()
	delete(wsListeners, conn)
}

const (
	debug   = "debug"
	error   = "error"
	info    = "info"
	warning = "warning"
)

// Debug logs a message with 'debug' level
func Debug(v ...interface{}) {
	log0(debug, fmt.Sprint(v...))
}

// Debugf logs a message with 'debug' level
func Debugf(format string, v ...interface{}) {
	log0(debug, fmt.Sprintf(format, v...))
}

// Info logs a message with 'info' level
func Info(v ...interface{}) {
	log0(info, fmt.Sprint(v...))
}

// Infof logs a message with 'info' level
func Infof(format string, v ...interface{}) {
	log0(info, fmt.Sprintf(format, v...))
}

// Error logs a message with 'error' level
func Error(v ...interface{}) {
	log0(error, fmt.Sprint(v...))
}

// Errorf logs a message with 'error' level
func Errorf(format string, v ...interface{}) {
	log0(error, fmt.Sprintf(format, v...))
}

// Warning logs a messages with 'warning' level
func Warning(v ...interface{}) {
	log0(warning, fmt.Sprint(v...))
}

// Warningf logs a message with 'warning' level
func Warningf(format string, v ...interface{}) {
	log0(warning, fmt.Sprintf(format, v...))
}

// Fatal is equivalent to Error() followed by a call to os.Exit(1)
func Fatal(v ...interface{}) {
	log0(error, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Errorf() followed by a call to os.Exit(1)
func Fatalf(format string, v ...interface{}) {
	log0(error, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func log0(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("%s %s: %s", timestamp, level, message)

	logEntry := protocol.LogEntry{Timestamp: timestamp, Level: level, Message: message}
	jsn, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("%s: couldn't marshal log entry: %v", error, err)
		return
	}

	for conn := range wsListeners {
		if err := conn.WriteMessage(websocket.TextMessage, jsn); err != nil {
			log.Printf("%s: couldn't log to websocket connection: %v", error, err)
			removeWSListener(conn)
		}
	}
}

// func SetConfig(applicationName, logger string) {
// 	if logger == "stderr" {
// 		// default logger
// 		log.SetPrefix(applicationName)
// 	} else if logger == "syslog" {
// 		writer, err := syslog.New(syslog.LOG_INFO, applicationName)
// 		if err != nil {
// 			log.Fatalf("Couldn't create logger: %v", err)
// 		}
// 		log.SetOutput(writer)
// 		log.SetFlags(0) // no timestamps etc, since syslog already prints that
// 	} else {
// 		f, err := os.OpenFile(logger, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
// 		if err != nil {
// 			log.Fatalf("Couldn't create logger: %v", err)
// 		}
// 		defer func() {
// 			err = f.Close()
// 			if err != nil {
// 				log.Fatalf("Couldn't close logger: %v", err)
// 			}
// 		}()
// 		log.SetOutput(f)
// 	}
// 	log.Printf("Created logger %v for %s", logger, applicationName)
// }
