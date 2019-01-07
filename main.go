package main

import (
	"fmt"
	"os"
	"os/exec"
)

var LogLevel string = "debug"

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
	Log("info", fmt.Sprintf("Dojo version %s", DojoVersion))

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
		PrintError(err.Error())
		os.Exit(1)
	}
	if mergedConfig.Debug == "true" {
		SetLogLevel("debug")
	} else {
		SetLogLevel("info")
	}
	Log("debug", fmt.Sprint("Config verified successfully"))
	Log("debug", fmt.Sprintf("Action set to: %s", mergedConfig.Action))
	Log("debug", fmt.Sprintf("ConfigFile set to: %s", mergedConfig.ConfigFile))
	Log("debug", fmt.Sprintf("Driver set to: %s", mergedConfig.Driver))
	Log("debug", fmt.Sprintf("Debug set to: %s", mergedConfig.Debug))
	Log("debug", fmt.Sprintf("Interactive set to: %s", mergedConfig.Interactive))
	Log("debug", fmt.Sprintf("RemoveContainers set to: %s", mergedConfig.RemoveContainers))
	Log("debug", fmt.Sprintf("WorkDirInner set to: %s", mergedConfig.WorkDirInner))
	Log("debug", fmt.Sprintf("WorkDirOuter set to: %s", mergedConfig.WorkDirOuter))
	Log("debug", fmt.Sprintf("IdentityDirOuter set to: %s", mergedConfig.IdentityDirOuter))
	Log("debug", fmt.Sprintf("Interactive set to: %s", mergedConfig.Interactive))
	Log("debug", fmt.Sprintf("BlacklistVariables set to: %s", mergedConfig.BlacklistVariables))
	Log("debug", fmt.Sprintf("DockerRunCommand set to: %s", mergedConfig.DockerRunCommand))
	Log("debug", fmt.Sprintf("DockerImage set to: %s", mergedConfig.DockerImage))
	Log("debug", fmt.Sprintf("DockerOptions set to: %s", mergedConfig.DockerOptions))
	Log("debug", fmt.Sprintf("DockerComposeFile set to: %s", mergedConfig.DockerComposeFile))

	//cmdString := "docker run -ti alpine:3.8"
	//RunShellInteractive(cmdString)
}
