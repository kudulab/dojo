package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var LogLevel string = "debug"

func handleConfig() Config {
	configFromCLI:= getCLIConfig()
	// debug option can be set on CLI only
	if configFromCLI.Debug == "true" {
		SetLogLevel("debug")
	} else {
		SetLogLevel("info")
	}
	configFile := configFromCLI.ConfigFile
	if configFile == "" {
		configFile = "Dojofile"
	}
	configFromFile := getFileConfig(configFile)
	defaultConfig := getDefaultConfig(configFile)
	mergedConfig := getMergedConfig(configFromCLI, configFromFile, defaultConfig)
	err := verifyConfig(&mergedConfig)
	if err != nil {
		Log("error", err.Error())
		os.Exit(1)
	}
	Log("debug", fmt.Sprintf("configFromCLI: %s", configFromCLI))
	Log("debug", fmt.Sprintf("configFromFile: %s", configFromFile))
	Log("debug", fmt.Sprintf("mergedConfig: %s", mergedConfig))
	Log("debug", fmt.Sprint("Config verified successfully"))
	return mergedConfig
}

func handleRun(mergedConfig Config, runID string) int {
	exitStatus := 0
	envService := EnvService{}
	if envService.IsCurrentUserRoot() {
		Log("warn", "Current user is root, which is not recommended")
	}
	fileService := FileService{}
	envFile := getEnvFilePath(runID, mergedConfig.Test)
	saveEnvToFile(fileService, envFile, mergedConfig.BlacklistVariables, envService.Variables())
	defer fileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFile)

	shellService := NewBashShellService()

	if mergedConfig.Driver == "docker"{
		dockerDriver := NewDockerDriver(shellService, fileService)
		exitStatus = dockerDriver.HandleRun(mergedConfig, runID, envFile)
	} else {
		// driver: docker-compose
		dcDriver := NewDockerComposeDriver(shellService, fileService)
		exitStatus = dcDriver.HandleRun(mergedConfig, runID, envFile)
	}
	return exitStatus
}


func handlePull(mergedConfig Config) int {
	exitStatus := 0
	if mergedConfig.Driver == "docker" {
		d := NewDockerDriver(NewBashShellService(), FileService{})
		exitStatus = d.HandlePull(mergedConfig)
	}
	return exitStatus
}

func handleCleanupOnSignal(mergedConfig Config, runID string) int {
	var exitStatus int
	if mergedConfig.Action == "run" {
		if runID == "" {
			Log("debug", "No cleaning needed")
			return 0
		}
		if mergedConfig.Driver == "docker" {
			d := NewDockerDriver(NewBashShellService(), FileService{})
			exitStatus = d.HandleSignal(mergedConfig, runID)
		} else {
			// driver: docker-compose
			d := NewDockerComposeDriver(NewBashShellService(), FileService{})
			exitStatus = d.HandleSignal(mergedConfig, runID)
		}
	}
	return exitStatus
}

func main() {
	Log("info", fmt.Sprintf("Dojo version %s", DojoVersion))
	mergedConfig := handleConfig()

	// This variable is needed to perform cleanup on any signal.
	// In order to avoid race conditions, let's write to this variable before
	// using multiple goroutines. And let's never write to it again.
	runID := ""
	if mergedConfig.Action == "run" {
		runID = getRunID(mergedConfig.Test)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	doneChannel := make(chan int, 1)

	go func(){
		if mergedConfig.Action == "run" {
			exitstatus := handleRun(mergedConfig, runID)
			doneChannel <- exitstatus
		} else if mergedConfig.Action == "pull" {
			exitstatus := handlePull(mergedConfig)
			doneChannel <- exitstatus
		}
	}()

	for {
		select {
		case signalCaught := <-signalChannel:
			Log("error", fmt.Sprintf("Caught signal: %s", signalCaught.String()))
			exitStatus := 1
			switch signalCaught {
			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				exitStatus = 130
				// kill -SIGTERM XXXX, GoCD uses it to cancel tasks
			case syscall.SIGTERM:
				exitStatus = 2
			}
			fmt.Println(exitStatus)
			exitStatusFromCleanup := handleCleanupOnSignal(mergedConfig, runID)
			if exitStatusFromCleanup != 0 {
				os.Exit(exitStatusFromCleanup)
			}
			os.Exit(exitStatus)
		case done := <-doneChannel:
			Log("debug", "Done, normal way")
			os.Exit(done)
		}
	}
}
