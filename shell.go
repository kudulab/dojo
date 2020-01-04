package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
	go_ps "github.com/mitchellh/go-ps"
	gopsutil "github.com/shirou/gopsutil/process"
)

type ShellServiceInterface interface {
	// Returns: exitStatus int and signaled bool.
	// Set separatePGroup to true in order to ignore signals. Then, you should never
	// process the signaled return value.
	RunInteractive(cmdString string, separatePGroup bool) (int, bool)
	// Returns: stdout string, stderr string, exitStatus int and signaled bool.
	// Set separatePGroup to true in order to ignore signals. Then, you should never
	// process the signaled return value.
	RunGetOutput(cmdString string, separatePGroup bool) (string, string, int, bool)
	CheckIfInteractive() bool
	// set environment variables, override any existing variables
	SetEnvironment(variables []string)
	GetProcessPid(processStrings []string) int
}

func NewBashShellService(logger *Logger) *BashShellService {
	return &BashShellService{
		Logger: logger,
		Environment: make([]string, 0),
	}
}

type BashShellService struct {
	Logger *Logger
	// collection of environment variables,
	// which are supposed to be preserved when running
	// a command with this struct
	Environment []string
}
func (bs BashShellService) GetProcessPid(processStrings []string) int {
	allProcesses, err := go_ps.Processes()
	if err != nil {
		panic(err)
	}
	for _, v := range allProcesses {
		processPid := int32(v.Pid())
		theProcess := gopsutil.Process{Pid: processPid}
		processCmd, cmdErr := theProcess.Cmdline()
		if cmdErr != nil {
			return -1
		}
		processName := processCmd
		allStringsContained := true
		for _, stringToBeContained := range processStrings {
			if ! strings.Contains(processName, stringToBeContained) {
				allStringsContained = false
				break
			}
		}
		if allStringsContained {
			return int(v.Pid())
		}
	}
	return -1
}

func (bs *BashShellService) SetEnvironment(variables []string) {
	bs.Environment = make([]string, 0)
	for _, value := range variables {
		bs.Environment = append(bs.Environment, value)
	}
}

func (bs BashShellService) RunInteractive(cmdString string, separatePGroup bool) (int, bool) {
	cmd := exec.Command("bash", "-c", cmdString)
	if separatePGroup {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			// Run in a separate process group, so that signals are not preserved. In theory
			// Setpgid: true should work, but it does not. Maybe this is because we run in "bash -c" ?
			// https://stackoverflow.com/questions/43364958/start-command-with-new-process-group-id-golang
			Setsid: true,
		}
	}
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Env = bs.Environment

	err := cmd.Run()

	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitStatus := status.ExitStatus()
	signaled := status.Signaled()
	signal := status.Signal()
	if err != nil && exitStatus == 0 {
		panic(fmt.Sprintf("unexpected: err not nil, exitStatus was 0, while running: %s", cmdString))
	}
	if signaled {
		bs.Logger.Log("debug", fmt.Sprintf("Signal: %v, while running: %s", signal, cmdString))
	}
	return exitStatus, signaled
}

func (bs BashShellService) RunGetOutput(cmdString string, separatePGroup bool) (string, string, int, bool) {
	cmd := exec.Command("bash", "-c", cmdString)
	if separatePGroup {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			// Run in a separate process group, so that signals are not preserved. In theory
			// Setpgid: true should work, but it does not. Maybe this is because we run in "bash -c" ?
			// https://stackoverflow.com/questions/43364958/start-command-with-new-process-group-id-golang
			Setsid: true,
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = bs.Environment

	err := cmd.Run()

	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitStatus := status.ExitStatus()
	signaled := status.Signaled()
	signal := status.Signal()
	if err != nil && exitStatus == 0 {
		panic(fmt.Sprintf("unexpected: err not nil, exitStatus was 0, while running: %s", cmdString))
	}
	if signaled {
		bs.Logger.Log("debug", fmt.Sprintf("Signal: %v, while running: %s", signal, cmdString))
	}
	return stdout.String(), stderr.String(), exitStatus, signaled
}

func (bs BashShellService) CheckIfInteractive() bool {
	// stolen from: https://github.com/mattn/go-isatty/blob/master/isatty_linux.go
	fd := os.Stdout.Fd()

	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(ioctlReadTermios), uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	interactive := (err == 0)
	bs.Logger.Log("debug", fmt.Sprintf("Current shell is interactive: %v", interactive))
	return interactive
}
