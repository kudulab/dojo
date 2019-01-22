package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)


type MockedShellServiceNotInteractive struct {
	ShellBinary string
	Logger *Logger
	CommandsReactions map[string]interface{}
}
func NewMockedShellServiceNotInteractive(logger *Logger) MockedShellServiceNotInteractive {
	return MockedShellServiceNotInteractive{
		Logger: logger,
	}
}
func NewMockedShellServiceNotInteractive2(logger *Logger, commandsReactions map[string]interface{}) MockedShellServiceNotInteractive {
	return MockedShellServiceNotInteractive{
		Logger: logger,
		CommandsReactions: commandsReactions,
	}
}
func (bs MockedShellServiceNotInteractive) RunInteractive(cmdString string, separePGroup bool) (int, bool) {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return 0, false
}
func (bs MockedShellServiceNotInteractive) RunGetOutput(cmdString string, separePGroup bool) (string, string, int, bool) {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
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
}
func NewMockedShellServiceInteractive(logger *Logger) MockedShellServiceInteractive {
	return MockedShellServiceInteractive{
		Logger: logger,
	}
}
func (bs MockedShellServiceInteractive) RunInteractive(cmdString string, separePGroup bool) (int, bool) {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return 0, false
}
func (bs MockedShellServiceInteractive) RunGetOutput(cmdString string, separePGroup bool) (string, string, int, bool) {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
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