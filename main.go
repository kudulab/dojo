package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)


func handleConfig(logger *Logger) Config {
	configFromCLI:= getCLIConfig()
	// debug option can be set on CLI only
	if configFromCLI.Debug == "true" {
		logger.SetLogLevel("debug")
	} else {
		logger.SetLogLevel("info")
	}
	configFile := configFromCLI.ConfigFile
	if configFile == "" {
		configFile = "Dojofile"
	} else {
		_, err := os.Lstat(configFile)
		if err != nil {
			if os.IsNotExist(err) {
				// user set custom config file and it does not exist
				logger.Log("error", fmt.Sprintf("ConfigFile set among cli options: \"%s\" does not exist", configFile))
				os.Exit(1)
			}
			panic(fmt.Sprintf("error when running os.Lstat(%q): %s", configFile, err))
		}
	}
	configFromFile := getFileConfig(logger, configFile)
	defaultConfig := getDefaultConfig(configFile)
	mergedConfig := getMergedConfig(configFromCLI, configFromFile, defaultConfig)
	err := verifyConfig(logger, &mergedConfig)
	if err != nil {
		logger.Log("error", err.Error())
		os.Exit(1)
	}
	logger.Log("debug", fmt.Sprintf("configFromCLI: %s", configFromCLI))
	logger.Log("debug", fmt.Sprintf("configFromFile: %s", configFromFile))
	logger.Log("debug", fmt.Sprintf("mergedConfig: %s", mergedConfig))
	logger.Log("debug", fmt.Sprint("Config verified successfully"))
	return mergedConfig
}

func handleSignal(logger *Logger, mergedConfig Config, runID string, driver DojoDriverInterface, multipleSignal bool) int {
	var exitStatus int
	if mergedConfig.Action != "run" {
		logger.Log("debug", "No cleaning needed (action was not: run)")
		return 0
	}
	if mergedConfig.Action == "run" {
		if runID == "" {
			logger.Log("debug", "No cleaning needed (runID not yet generated)")
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
	logger := NewLogger("debug")
	logger.Log("info", fmt.Sprintf("Dojo version %s", DojoVersion))
	mergedConfig := handleConfig(logger)

	fileService := NewFileService(logger)
	shellService := NewBashShellService(logger)
	var driver DojoDriverInterface
	if mergedConfig.Driver == "docker"{
		driver = NewDockerDriver(shellService, fileService, logger)
	} else {
		driver = NewDockerComposeDriver(shellService, fileService, logger)
	}

	if mergedConfig.Action == "pull" {
		exitstatus := driver.HandlePull(mergedConfig)
		os.Exit(exitstatus)
	}
	// action is run

	// This variable is needed to perform cleanup on any signal.
	// In order to avoid race conditions, let's write to this variable before
	// using multiple goroutines. And let's never write to it again.
	runID := getRunID(mergedConfig.Test)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	doneChannel := make(chan int, 1)

	go func(){
		envService := EnvService{}
		exitstatus := driver.HandleRun(mergedConfig, runID, envService)
		doneChannel <- exitstatus
	}()

	select {
	case signalCaught := <-signalChannel:
		logger.Log("error", fmt.Sprintf("Caught signal: %s", signalCaught.String()))
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
			exitStatusFromCleanup := handleSignal(logger, mergedConfig, runID, driver, false)
			if exitStatusFromCleanup != 0 {
				exitStatus = exitStatusFromCleanup
			}
			doneAfterOneSignalChannel <- exitStatus
		}()
		select {
		case multipleSignalCaught := <-multipleSignalChannel:
			logger.Log("error", fmt.Sprintf("Caught another signal: %s", multipleSignalCaught.String()))
			exitStatus = 3
			exitStatusFromCleanup := handleSignal(logger, mergedConfig, runID, driver, true)
			if exitStatusFromCleanup != 0 {
				exitStatus = exitStatusFromCleanup
			}
			logger.Log("debug", "Finished after multiple signals")
			os.Exit(exitStatus)

		case doneAfter1Signal := <- doneAfterOneSignalChannel:
			logger.Log("debug", "Finished after 1 signal")
			os.Exit(doneAfter1Signal)
		}
	case done := <-doneChannel:
		logger.Log("debug", "Finished, normal way")
		os.Exit(done)
	}
}
