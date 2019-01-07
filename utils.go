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
	if level != "info" && level != "debug" {
		panic(fmt.Sprintf("Unsupported log level: %v", level))
	}
	if level == "debug" && GetLogLevel() != "debug" {
		return // do not print debug log message
	}

	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	// info has 1 letter less than debug, so let's add a space in the beginning of a line
	prettyLogLevel := strings.ToUpper(level)
	if prettyLogLevel == "INFO" {
		prettyLogLevel = " INFO"
	}

	log.Printf("[%2d] %s: (%s) %s", getGoroutineID(), prettyLogLevel, frame.Function, msg)
}

func PrintError(msg string) {
	fullMsg := red(fmt.Sprintf("ERROR: %v\n", msg))
	fmt.Fprint(os.Stderr, fullMsg)
}

func red(text string) string {
	return "\033[31m" + text + "\033[0m"
}