package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os/exec"
	"strconv"
	"sync"
	"testing"
)


type MockedShellServiceNotInteractive struct {
	ShellBinary string
	Logger *Logger
	CommandsReactions map[string]interface{}
	CommandsRun []string
	Environment []string
	Mutex *sync.Mutex
}
func NewMockedShellServiceNotInteractive(logger *Logger) *MockedShellServiceNotInteractive {
	return &MockedShellServiceNotInteractive{
		Logger: logger,
		Mutex: &sync.Mutex{},
		CommandsRun: make([]string, 0),
	}
}
func (bs MockedShellServiceNotInteractive) SetEnvironment(variables []string) {
	bs.Environment = make([]string, 0)
	for _, value := range variables {
		bs.Environment = append(bs.Environment, value)
	}
}
func NewMockedShellServiceNotInteractive2(logger *Logger, commandsReactions map[string]interface{}) *MockedShellServiceNotInteractive {
	return &MockedShellServiceNotInteractive{
		Logger: logger,
		CommandsReactions: commandsReactions,
		Mutex: &sync.Mutex{},
		CommandsRun: make([]string, 0),
	}
}
func (ss *MockedShellServiceNotInteractive) AppendCommandRun(command string) {
	ss.Mutex.Lock()
	ss.CommandsRun = append(ss.CommandsRun, command)
	ss.Mutex.Unlock()
}

func (bs MockedShellServiceNotInteractive) GetProcessPid(processStrings []string) int {
	return -1
}

// AppendCommandRun needs to be invoked on a pointer to object, because it changes the object state.
// Since this method invokes AppendCommandRun() function, it also needs to be invoked on a pointer to object.
func (bs *MockedShellServiceNotInteractive) RunInteractive(cmdString string, separePGroup bool) (int, bool) {
	cmd := fmt.Sprintf("Pretending to run: %s", cmdString)
	bs.Logger.Log("debug", cmd)
	bs.AppendCommandRun(cmd)
	return 0, false
}
func (bs *MockedShellServiceNotInteractive) RunGetOutput(cmdString string, separePGroup bool) (string, string, int, bool) {
	cmd := fmt.Sprintf("Pretending to run: %s", cmdString)
	bs.Logger.Log("debug", cmd)
	bs.AppendCommandRun(cmd)
	if bs.CommandsReactions != nil {
		if val, ok := bs.CommandsReactions[cmdString]; ok {
			valArr := val.([]string)
			stdo := valArr[0]
			stde := valArr[1]
			es, err := strconv.Atoi(valArr[2])
			if err != nil {
				panic(err)
			}
			cmdInfo := cmdInfoToString(cmdString, stdo, stde, es)
			bs.Logger.Log("debug", fmt.Sprintf("Pretending to return: %s", cmdInfo))
			return stdo, stde, es, false
		}
	}
	return "", "", 0, false
}
func (bs MockedShellServiceNotInteractive) CheckIfInteractive() bool {
	return false
}

type MockedShellServiceInteractive struct {
	ShellBinary string
	Logger *Logger
	Environment []string
	CommandsRun []string
	Mutex *sync.Mutex
}
func (bs MockedShellServiceInteractive) SetEnvironment(variables []string) {
	bs.Environment = make([]string, 0)
	for _, value := range variables {
		bs.Environment = append(bs.Environment, value)
	}
}
func NewMockedShellServiceInteractive(logger *Logger) *MockedShellServiceInteractive {
	return &MockedShellServiceInteractive{
		Logger: logger,
		Mutex: &sync.Mutex{},
		CommandsRun: make([]string, 0),
	}
}
func (ss *MockedShellServiceInteractive) AppendCommandRun(command string) {
	ss.Mutex.Lock()
	ss.CommandsRun = append(ss.CommandsRun, command)
	ss.Mutex.Unlock()
}
func (bs *MockedShellServiceInteractive) RunInteractive(cmdString string, separePGroup bool) (int, bool) {
	cmd := fmt.Sprintf("Pretending to run: %s", cmdString)
	bs.Logger.Log("debug", cmd)
	bs.AppendCommandRun(cmd)
	return 0, false
}
func (bs *MockedShellServiceInteractive) RunGetOutput(cmdString string, separePGroup bool) (string, string, int, bool) {
	cmd := fmt.Sprintf("Pretending to run: %s", cmdString)
	bs.Logger.Log("debug", cmd)
	bs.AppendCommandRun(cmd)
	return "", "", 0, false
}
func (bs MockedShellServiceInteractive) CheckIfInteractive() bool {
	return true
}
func (bs MockedShellServiceInteractive) GetProcessPid(processStrings []string) int {
	return -1
}

func TestMockedShellService_CheckIfInteractive(t *testing.T){
	logger := NewLogger("debug")
	shell := NewMockedShellServiceNotInteractive(logger)
	interactive := shell.CheckIfInteractive()
	assert.False(t, interactive)
}
func TestBashShellService_CheckIfInteractive(t *testing.T) {
	logger := NewLogger("debug")
	shell := NewBashShellService(logger)
	interactive := shell.CheckIfInteractive()
	assert.False(t, interactive)
}
func TestBashShellService_RunInteractive(t *testing.T) {
	logger := NewLogger("debug")
	shell := NewBashShellService(logger)
	exitstatus, signaled := shell.RunInteractive("echo hello", false)
	assert.Equal(t, 0, exitstatus)
	assert.Equal(t, false, signaled)
}
func TestBashShellService_RunGetOutput(t *testing.T) {
	logger := NewLogger("debug")
	shell := NewBashShellService(logger)
	stdout, sterr, exitstatus, signaled := shell.RunGetOutput("echo hello",false)
	assert.Equal(t, "hello\n", stdout)
	assert.Equal(t, "", sterr)
	assert.Equal(t, 0, exitstatus)
	assert.Equal(t, false, signaled)
}
func TestBashShellService_SetEnv(t *testing.T) {
	logger := NewLogger("debug")
	shell := NewBashShellService(logger)
	shell.SetEnvironment([]string{"ABC=123", "DEF=444", "ZZZ=999", "YYY=666"})
	assert.Equal(t, 4, len(shell.Environment))
	assert.Equal(t, "ABC=123", shell.Environment[0])
}
func startSomeCmd(name, arg string) func() {
	var cmd = exec.Command(name, arg)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	// Return a function to be invoked with "defer", so that the process is killed
	// even if tests are failed. Idea from: https://stackoverflow.com/a/42310257/4457564
	return func() { cmd.Process.Kill() }
}

func TestBashShellService_GetProcessPid(t *testing.T) {
	var processTests = []struct {
		names []string
		proccessFound bool
	} {
		{[]string{"sleep", "5"}, true},
		{[]string{"sleep", "50"}, true},
		{[]string{"sleep"}, true},
		{[]string{"sleep", "1"}, false},
		{[]string{"blaaa"}, false},
	}

	deferFunc1 := startSomeCmd("sleep", "5")
	deferFunc2 := startSomeCmd("sleep", "50")
	defer deferFunc1()
	defer deferFunc2()
	for _, tt := range processTests {
		logger := NewLogger("debug")
		shell := NewBashShellService(logger)
		pid := shell.GetProcessPid(tt.names)
		if tt.proccessFound {
			assert.True(t, pid > 0, tt.names)
		} else {
			assert.Equal(t, pid, -1, tt.names)
		}
	}
}