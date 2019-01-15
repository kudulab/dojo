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
	} else {
		_, err := os.Lstat(configFile)
		if err != nil {
			if os.IsNotExist(err) {
				// user set custom config file and it does not exist
				Log("error", fmt.Sprintf("ConfigFile set among cli options: \"%s\" does not exist", configFile))
				os.Exit(1)
			}
			panic(fmt.Sprintf("error when running os.Lstat(%q): %s", configFile, err))
		}
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

func handleSignal(mergedConfig Config, runID string, driver DojoDriverInterface, multipleSignal bool) int {
	var exitStatus int
	if mergedConfig.Action == "pull" {
		Log("debug", "No cleaning needed (pull action)")
		return 0
	}
	if mergedConfig.Action == "run" {
		if runID == "" {
			Log("debug", "No cleaning needed (runID not yet generated)")
			return 0
		}
		if multipleSignal {
			exitStatus = driver.HandleMultipleSignal(mergedConfig, runID)
		} else {
			exitStatus = driver.HandleSignal(mergedConfig, runID)
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

	envService := EnvService{}
	fileService := FileService{}
	shellService := NewBashShellService()
	var driver DojoDriverInterface
	if mergedConfig.Driver == "docker"{
		driver = NewDockerDriver(shellService, fileService)
	} else {
		driver = NewDockerComposeDriver(shellService, fileService)
	}

	go func(){
		if mergedConfig.Action == "run" {
			exitstatus := driver.HandleRun(mergedConfig, runID, envService)
			doneChannel <- exitstatus
		} else if mergedConfig.Action == "pull" {
			exitstatus := driver.HandlePull(mergedConfig)
			doneChannel <- exitstatus
		}
	}()

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

		multipleSignalChannel := make(chan os.Signal, 1)
		signal.Notify(multipleSignalChannel, os.Interrupt, syscall.SIGTERM)
		doneAfterOneSignalChannel := make(chan int, 1)
		go func() {
			exitStatusFromCleanup := handleSignal(mergedConfig, runID, driver, false)
			if exitStatusFromCleanup != 0 {
				exitStatus = exitStatusFromCleanup
			}
			doneAfterOneSignalChannel <- exitStatus
		}()
		select {
		case multipleSignalCaught := <-multipleSignalChannel:
			Log("error", fmt.Sprintf("Caught another signal: %s", multipleSignalCaught.String()))
			exitStatus = 3
			exitStatusFromCleanup := handleSignal(mergedConfig, runID, driver, true)
			if exitStatusFromCleanup != 0 {
				exitStatus = exitStatusFromCleanup
			}
			Log("debug", "Finished after multiple signals")
			os.Exit(exitStatus)

		case doneAfter1Signal := <- doneAfterOneSignalChannel:
			Log("debug", "Finished after 1 signal")
			os.Exit(doneAfter1Signal)
		}
	case done := <-doneChannel:
		Log("debug", "Finished, normal way")
		os.Exit(done)
	}
}
