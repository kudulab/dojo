package main

import (
	"os"
	"os/exec"
)

func RunShellInteractive(cmdString string) {
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	println(err)

	exitStatus := 0
	println(exitStatus)
}

func main() {
	println("hello in dojo")
	cmdString := "docker run -ti alpine:3.8"
	RunShellInteractive(cmdString)
}