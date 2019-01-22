package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

type ShellServiceInterface interface {
	RunInteractive(cmdString string) int
	RunGetOutput(cmdString string) (string, string, int)
	CheckIfInteractive() bool
}

func NewBashShellService(logger *Logger) BashShellService {
	return BashShellService{
		Logger: logger,
	}
}

type BashShellService struct {
	Logger *Logger
}


func (bs BashShellService) RunInteractive(cmdString string) int {
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitStatus := status.ExitStatus()
	signaled := status.Signaled()
	signal := status.Signal()
	if err != nil && exitStatus ==0 {
		panic("unexpected: err not nil, exitStatus was 0")
	}
	if signaled {
		bs.Logger.Log("debug", fmt.Sprintf("Signal: %v", signal))
	}
	return exitStatus
}

func (bs BashShellService) RunGetOutput(cmdString string) (string, string, int) {
	cmd := exec.Command("bash", "-c", cmdString)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitStatus := status.ExitStatus()
	signaled := status.Signaled()
	signal := status.Signal()
	if err != nil && exitStatus ==0 {
		panic("unexpected: err not nil, exitStatus was 0")
	}
	if signaled {
		bs.Logger.Log("debug", fmt.Sprintf("Signal: %v", signal))
	}
	return stdout.String(), stderr.String(), exitStatus
}

func (bs BashShellService) CheckIfInteractive() bool {
	// stolen from: https://github.com/mattn/go-isatty/blob/master/isatty_linux.go
	fd := os.Stdout.Fd()
	const ioctlReadTermios = syscall.TCGETS

	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	interactive := (err == 0)
	bs.Logger.Log("debug", fmt.Sprintf("Current shell is interactive: %v", interactive))
	return interactive
}
