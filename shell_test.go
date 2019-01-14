package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)


type MockedShellServiceNotInteractive struct {
	ShellBinary string
}
func (bs MockedShellServiceNotInteractive) RunInteractive(cmdString string) int {
	Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return 0
}
func (bs MockedShellServiceNotInteractive) RunGetOutput(cmdString string) (string, string, int) {
	Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return "", "", 0
}
func (bs MockedShellServiceNotInteractive) CheckIfInteractive() bool {
	return false
}

type MockedShellServiceInteractive struct {
	ShellBinary string
}
func (bs MockedShellServiceInteractive) RunInteractive(cmdString string) int {
	Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return 0
}
func (bs MockedShellServiceInteractive) RunGetOutput(cmdString string) (string, string, int) {
	Log("debug", fmt.Sprintf("Pretending to run: %s", cmdString))
	return "", "", 0
}
func (bs MockedShellServiceInteractive) CheckIfInteractive() bool {
	return true
}

func TestMockedShellService_CheckIfInteractive(t *testing.T){
	shell := MockedShellServiceNotInteractive{}
	interactive := shell.CheckIfInteractive()
	assert.False(t, interactive)
}
func TestBashShellService_CheckIfInteractive(t *testing.T) {
	shell := NewBashShellService()
	interactive := shell.CheckIfInteractive()
	assert.False(t, interactive)
}
func TestBashShellService_RunInteractive(t *testing.T) {
	shell := NewBashShellService()
	exitstatus := shell.RunInteractive("echo hello")
	assert.Equal(t, 0, exitstatus)
}
func TestBashShellService_RunGetOutput(t *testing.T) {
	shell := NewBashShellService()
	stdout, sterr, exitstatus := shell.RunGetOutput("echo hello")
	assert.Equal(t, "hello\n", stdout)
	assert.Equal(t, "", sterr)
	assert.Equal(t, 0, exitstatus)
}