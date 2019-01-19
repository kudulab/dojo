package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
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
func green(text string) string {
	return "\033[32m" + text + "\033[0m"
}
func orange(text string) string {
	return "\033[33m" + text + "\033[0m"
}

// Returns an identificator that can be reused later in many places,
// e.g. as some file name or as docker container name.
// e.g. dojo-myproject-2019-01-09_10-39-06-98498093
// It may not contain upper case letters or else "docker inspect" complains with the error:
// invalid reference format: repository name must be lowercase.
func getRunID(test string) string {
	if test != "true" {
		currentDirectory, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		currentDirectorySplit := strings.Split(currentDirectory, "/")
		currentDirectoryLastPart := currentDirectorySplit[len(currentDirectorySplit)-1]

		currentTime := time.Now().Format("2006-01-02_15-04-05")
		// run ID must contain a random number. Using time is insufficient, because e.g. 2 CI agents may be started
		// in the same second for the same project.
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(99999999)
		return fmt.Sprintf("dojo-%s-%v-%v", currentDirectoryLastPart, currentTime, randomNumber)
	} else {
		return "testdojorunid"
	}
}

func cmdInfoToString(cmd string, stdout string, stderr string, exitStatus int) string {
	return fmt.Sprintf(`Command: %s
exit status: %v
stdout: %s
stderr: %s`,
		cmd, exitStatus, stdout, stderr)
}