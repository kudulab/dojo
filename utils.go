package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func GetLogLevel() string {
	return LogLevel
}
func SetLogLevel(level string) {
	LogLevel = strings.ToLower(level)
}

func Log(level, msg string) {
	if level != "info" && level != "debug" && level != "warn" && level != "error" {
		panic(fmt.Sprintf("Unsupported log level: %v", level))
	}
	if level == "debug" && GetLogLevel() != "debug" {
		return // do not print debug log message
	}

	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	coloredMsg := msg

	// info has 1 letter less than debug, so let's add a space in the beginning of a line
	prettyLogLevel := strings.ToUpper(level)
	if prettyLogLevel == "INFO" {
		prettyLogLevel = " INFO"
	}
	if prettyLogLevel == "WARN" {
		prettyLogLevel = " WARN"
		coloredMsg = orange(msg)
	}
	if prettyLogLevel == "ERROR" {
		coloredMsg = red(msg)
	}

	log.Printf("[%2d] %s: (%s) %s", getGoroutineID(), prettyLogLevel, frame.Function, coloredMsg)
}

func red(text string) string {
	return "\033[31m" + text + "\033[0m"
}
func orange(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func removeFile(filePath string, ignoreNoSuchFileError bool) {
	err := os.Remove(filePath)
	if err != nil {
		if strings.Contains(err.Error(),"no such file or directory") && ignoreNoSuchFileError {
			return
		}
		panic(err)
	}
}