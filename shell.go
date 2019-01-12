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

func NewBashShellService() BashShellService {
	return BashShellService{"bash"}
}

type BashShellService struct {
	ShellBinary string
}

func (bs BashShellService) RunInteractive(cmdString string) int {
	cmd := exec.Command("bash", "-c", cmdString)
	// we may want to experiment more with:
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	//	Setpgid: true,
	//	//  Process Group ID; The unique positive integer identifier representing a process group during its lifetime.
	//	//  The child processes started here (in golang) are starting in the same process group as the creator-process (parent) by default.
	// //  (Bash starts each child process in their own process group).
	// //  When a signal is directed to a process group, the signal is delivered to each process that is a member of the group.
	//	Pgid:    0,
	//}
	// because running "docker run -ti" and then ctrl+c sometimes results in a frozen terminal.
	// This has nothing to do with golang, but we may want to try to fix it somehow. However,
	// we cannot just kill the "docker run" command in the last select block, because when should we do it?
	// Signal is not caught then.
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitStatus := status.ExitStatus()
	signaled := status.Signaled()
	signal := status.Signal()
	if err != nil {
		Log("error", fmt.Sprintf("err: %v", err))
	}
	if signaled {
		Log("error", fmt.Sprintf("Signal: %v", signal))
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

	exitStatus := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
				return stdout.String(), stderr.String(), exitStatus
			}
		}
		return stdout.String(), stderr.String(), 1
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
	Log("debug", fmt.Sprintf("Current shell is interactive: %v", interactive))
	return interactive
}
