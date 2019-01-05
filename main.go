package main

import (
	"fmt"
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
	configFromCLI:= getCLIConfig()
	configFile := configFromCLI.ConfigFile
	if configFile == "" {
		configFile = "Dojofile"
	}
	configFromFile := getFileConfig(configFile)
	defaultConfig := getDefaultConfig(configFile)
	mergedConfig := getMergedConfig(configFromCLI, configFromFile, defaultConfig)
	err := verifyConfig(mergedConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Dojo error: %v\n", err)
		os.Exit(1)
	}
	println("hello in dojo")
	cmdString := "docker run -ti alpine:3.8"
	RunShellInteractive(cmdString)
}
