package main

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

func RunShell(cmdString string) int {
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	exitStatus := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
				return exitStatus
			}
		}
		return 1
	}
	return exitStatus
}

func checkIfInteractive() bool {
	// stolen from: https://github.com/mattn/go-isatty/blob/master/isatty_linux.go
	fd := os.Stdout.Fd()
	const ioctlReadTermios = syscall.TCGETS

	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
