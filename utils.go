package main

import (
	"bytes"
	"fmt"
	"hash/maphash"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Logger struct {
	Level string
}

func NewLogger(level string) *Logger {
	return &Logger{
		Level: strings.ToLower(level),
	}
}

func (l *Logger) SetLogLevel(level string) {
	l.Level = strings.ToLower(level)
}

func (l *Logger) SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func (l *Logger) Log(level, msg string) {
	if level != "info" && level != "debug" && level != "warn" && level != "error" {
		panic(fmt.Sprintf("Unsupported log level: %v", level))
	}
	if level == "debug" && l.Level != "debug" {
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
		return getRunIDGenerateFromCurrentDir(currentDirectory)
	} else {
		return "testdojorunid"
	}
}

func getRunIDGenerateFromCurrentDir(currentDirectory string) string {
	currentDirectorySplit := strings.Split(currentDirectory, "/")
	currentDirectoryLastPart := currentDirectorySplit[len(currentDirectorySplit)-1]

	currentTime := time.Now().Format("2006-01-02_15-04-05")
	// run ID must contain a random number. Using time is insufficient, because e.g. 2 CI agents may be started
	// in the same second for the same project.
	// https://stackoverflow.com/a/73251027
	r := rand.New(rand.NewSource(int64(new(maphash.Hash).Sum64())))
	randomNumber := r.Int()
	runID := fmt.Sprintf("dojo-%s-%v-%v", currentDirectoryLastPart, currentTime, randomNumber)
	// replace all the upper case letters to lower case letters, this is to support the case when
	// the currentDirectory contains capital letters and docker-compose project names do not welcome
	// capital letters
	runID = strings.ToLower(runID)
	//runID = strings.ReplaceAll(runID, " ", "")
	// remove any special characters
	runID = regexp.MustCompile(`[^a-zA-Z0-9\-]+`).ReplaceAllString(runID, "")
	return runID
}

func cmdInfoToString(cmd string, stdout string, stderr string, exitStatus int) string {
	if stdout == "" {
		stdout = "<empty string>"
	} else {
		stdout = strings.TrimSuffix(stdout, "\n")
	}
	if stderr == "" {
		stderr = "<empty string>"
	} else {
		stderr = strings.TrimSuffix(stderr, "\n")
	}
	return fmt.Sprintf(`Command: %s
  Exit status: %v
  StdOut: %s
  StdErr: %s
`,
		cmd, exitStatus, stdout, stderr)
}

func removeWhiteSpaces(str string) string {
	return strings.Join(strings.Fields(str), "")
}

type ContainerInfo struct {
	ID       string
	Name     string
	Status   string
	ExitCode string
	Exists   bool
	Logs     string
}

// Returns: container ID, status , whether or not a container exists, error
func getContainerInfo(shellService ShellServiceInterface, containerNameOrID string) (*ContainerInfo, error) {
	if containerNameOrID == "" {
		panic("containerNameOrID was empty")
	}
	// https://docs.docker.com/engine/api/v1.21/
	cmd := fmt.Sprintf("docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' %s", containerNameOrID)
	stdout, stderr, exitStatus, _ := shellService.RunGetOutput(cmd, true)
	if exitStatus != 0 {
		if strings.Contains(stdout, "No such object") || strings.Contains(stderr, "No such object") {
			return &ContainerInfo{Exists: false}, nil
		}
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
		return &ContainerInfo{}, fmt.Errorf("Unexpected exit status:\n%s", cmdInfo)
	}
	status := strings.TrimSuffix(stdout, "\n")
	outputArr := strings.Split(status, " ")
	return &ContainerInfo{
		ID:       outputArr[0],
		Name:     strings.TrimPrefix(outputArr[1], "/"),
		Status:   outputArr[2],
		ExitCode: outputArr[3],
		Exists:   true,
	}, nil
}
