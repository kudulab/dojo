package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)


type MockedShellServiceNotInteractive struct {
	ShellBinary string
	Logger *Logger
}
func NewMockedShellServiceNotInteractive(logger *Logger) MockedShellServiceNotInteractive {
	return MockedShellServiceNotInteractive{
		Logger: logger,
	}
}
func (bs MockedShellServiceNotInteractive) RunInteractive(cmdString string) int {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return 0
}
func (bs MockedShellServiceNotInteractive) RunGetOutput(cmdString string) (string, string, int) {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return "", "", 0
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
func (bs MockedShellServiceInteractive) RunInteractive(cmdString string) int {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return 0
}
func (bs MockedShellServiceInteractive) RunGetOutput(cmdString string) (string, string, int) {
	bs.Logger.Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return "", "", 0
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
	exitstatus := shell.RunInteractive("echo hello")
	assert.Equal(t, 0, exitstatus)
}
func TestBashShellService_RunGetOutput(t *testing.T) {
	logger := NewLogger("debug")
	shell := NewBashShellService(logger)
	stdout, sterr, exitstatus := shell.RunGetOutput("echo hello")
	assert.Equal(t, "hello\n", stdout)
	assert.Equal(t, "", sterr)
	assert.Equal(t, 0, exitstatus)
}