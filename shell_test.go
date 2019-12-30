package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
}
func NewMockedShellServiceNotInteractive(logger *Logger) *MockedShellServiceNotInteractive {
	return &MockedShellServiceNotInteractive{
		Logger: logger,
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
	}
}
func (ss *MockedShellServiceNotInteractive) AppendCommandRun(command string) {
	if ss.CommandsRun == nil {
		ss.CommandsRun = make([]string, 0)
	}
	ss.CommandsRun = append(ss.CommandsRun, command)
}
func (bs MockedShellServiceNotInteractive) RunInteractive(cmdString string, separePGroup bool) (int, bool) {
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
	}
}
func (ss *MockedShellServiceInteractive) AppendCommandRun(command string) {
	if ss.CommandsRun == nil {
		ss.Mutex = &sync.Mutex{}
		ss.CommandsRun = make([]string, 0)
	}
	ss.Mutex.Lock()
	ss.CommandsRun = append(ss.CommandsRun, command)
	ss.Mutex.Unlock()
}
func (bs MockedShellServiceInteractive) RunInteractive(cmdString string, separePGroup bool) (int, bool) {
	cmd := fmt.Sprintf("Pretending to run: %s", cmdString)
	bs.Logger.Log("debug", cmd)
	bs.AppendCommandRun(cmd)
	return 0, false
}
func (bs MockedShellServiceInteractive) RunGetOutput(cmdString string, separePGroup bool) (string, string, int, bool) {
	cmd := fmt.Sprintf("Pretending to run: %s", cmdString)
	bs.Logger.Log("debug", cmd)
	bs.AppendCommandRun(cmd)
	return "", "", 0, false
}
func (bs MockedShellServiceInteractive) CheckIfInteractive() bool {
	return true
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
