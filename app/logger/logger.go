package logger

import (
	"log"
	"os"
	"runtime"
)

func Info(message string) {
	_, file, line, _ := runtime.Caller(1)
	infoLogger := log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	infoLogger.Printf("%s:%d - %s\n", file, line, message)
}

func Error(message string) {
	_, file, line, _ := runtime.Caller(1)
	errorLogger := log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
	errorLogger.Printf("%s:%d - %s\n", file, line, message)
}
