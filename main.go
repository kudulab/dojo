package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
)

var LogLevel string = "debug"

func handleConfig() Config {
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
	Log("debug", fmt.Sprintf("configFromCLI: %s", configFromCLI))
	Log("debug", fmt.Sprintf("configFromFile: %s", configFromFile))
	Log("debug", fmt.Sprintf("mergedConfig: %s", mergedConfig))
	Log("debug", fmt.Sprint("Config verified successfully"))
	return mergedConfig
}

func handleRun(mergedConfig Config) int {
	exitStatus := 0
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	if currentUser.Username == "root" {
		Log("warn", "Current user is root, which is not recommended")
	}

	if mergedConfig.Driver == "docker"{
		runID := getRunID()
		envFile := fmt.Sprintf("/tmp/dojo-environment-%s", runID)
		saveEnvToFile(envFile, mergedConfig.BlacklistVariables, mergedConfig.Dryrun)
		Log("debug", fmt.Sprintf("Saved environment variables to file: %v", envFile))
		interactiveShell := checkIfInteractive()
		Log("debug", fmt.Sprintf("Current shell is interactive: %v", interactiveShell))
		cmd := constructDockerCommand(mergedConfig, envFile, runID, interactiveShell)
		Log("info", fmt.Sprintf("docker command will be:\n %v", cmd))
		if mergedConfig.RemoveContainers != "true" && mergedConfig.Dryrun != "true" {
			// Removing docker container is impractical without additional steps here. We'd have to
			// parse output of dojo in order to get container name. Thus, save the container name to a file.
			currentDirectory, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			rcFile := fmt.Sprintf("%s/dojorc.txt", currentDirectory)
			removeFile(rcFile, true)
			err1 := ioutil.WriteFile(rcFile, []byte(runID), 0644)
			if err1 != nil {
				panic(err1)
			}
			Log("info", fmt.Sprintf("Written docker container name to file %s", rcFile))

			rcFile2 := fmt.Sprintf("%s/dojorc", currentDirectory)
			removeFile(rcFile2, true)
			err2 := ioutil.WriteFile(rcFile2, []byte(fmt.Sprintf("DOJO_RUN_ID=%s",runID)), 0644)
			if err2 != nil {
				panic(err2)
			}
			Log("info", fmt.Sprintf("Written docker container name to file %s", rcFile2))
		}
		if mergedConfig.Dryrun != "true" {
			exitStatus = RunShell(cmd)
			Log("debug", fmt.Sprintf("Exit status: %v", exitStatus))
		} else {
			Log("info", "Dryrun set, not running docker container")
		}
		if mergedConfig.Dryrun != "true" {
			os.Remove(envFile)
		}
	} else {
		// driver: docker-compose
	}
	return exitStatus
}
func handlePull(mergedConfig Config) int {
	exitStatus := 0
	if mergedConfig.Driver == "docker" {
		cmd := fmt.Sprintf("docker pull %s", mergedConfig.DockerImage)
		if mergedConfig.Dryrun != "true" {
			exitStatus = RunShell(cmd)
			Log("debug", fmt.Sprintf("Exit status: %v", exitStatus))
		} else {
			Log("info", "Dryrun set, not pulling docker image")
		}
	}
	return exitStatus
}

func main() {
	Log("info", fmt.Sprintf("Dojo version %s", DojoVersion))
	mergedConfig := handleConfig()

	if mergedConfig.Action == "run" {
		exitstatus := handleRun(mergedConfig)
		os.Exit(exitstatus)
	} else if mergedConfig.Action == "pull" {
		exitstatus := handlePull(mergedConfig)
		os.Exit(exitstatus)
	}
}
