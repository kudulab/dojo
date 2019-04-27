package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
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

	envService := EnvService{}
	newVariables := []string{
		fmt.Sprintf("DOJO_WORK_INNER=%s", mergedConfig.WorkDirInner),
		fmt.Sprintf("DOJO_WORK_OUTER=%s", mergedConfig.WorkDirOuter)}
	shellService.SetEnvironment(envService.Variables(), newVariables)

	if mergedConfig.Action == "pull" {
		exitstatus := driver.HandlePull(mergedConfig)
		os.Exit(exitstatus)
	}

	// action is run

	// This variable is needed to perform cleanup on any signal.
	// In order to avoid race conditions, let's write to this variable before
	// using multiple goroutines. And let's never write to it again.
	runID := getRunID(mergedConfig.Test)
	doneChannel := make(chan int, 1)
	signalChannel := registerSignalChannel()

	// main work goroutine
	go func(){
		// run and stop the containers
		exitstatus := driver.HandleRun(mergedConfig, runID, envService)
		doneChannel <- exitstatus
	}()

	signalsCaughtCount := 0
	signalExitStatus := 0
	var wg sync.WaitGroup
	for {
		select {
		case signal := <-signalChannel:
			signalsCaughtCount++
			logger.Log("error", fmt.Sprintf("Caught signal %v: %s", signalsCaughtCount, signal.String()))
			if signalsCaughtCount == 1 {
				signalExitStatus = signalToExitStatus(signal)
				wg.Add(1)
				go func() {
					// the job here is to gracefully stop the main work
					handleSignal(logger, mergedConfig, runID, driver, false)
					wg.Done()
				}()
			} else if signalsCaughtCount == 2 {
				signalExitStatus = 3
				wg.Add(1)
				go func() {
					// the job here is to immediately stop the main work
					handleSignal(logger, mergedConfig, runID, driver, true)
					wg.Done()
				}()
			} else {
				logger.Log("debug", fmt.Sprintf("Ignoring signal %v: %s", signalsCaughtCount, signal.String()))
			}
		case done := <-doneChannel:
			logger.Log("debug", fmt.Sprintf("Finished main work"))
			exitStatus := done
			logger.Log("debug", fmt.Sprintf("Exit status from main work: %v", exitStatus))

			logger.Log("debug", fmt.Sprintf("Waiting for logic that handles signals"))
			wg.Wait()
			logger.Log("debug", fmt.Sprintf("Done waiting for logic that handles signals"))

			cleaningExitStatus := driver.CleanAfterRun(mergedConfig, runID)
			logger.Log("debug", fmt.Sprintf("Exit status from cleaning: %v", cleaningExitStatus))
			logger.Log("debug", fmt.Sprintf("Exit status from signals: %v", signalExitStatus))
			if cleaningExitStatus != 0 {
				exitStatus = cleaningExitStatus
			}
			if signalExitStatus != 0 {
				exitStatus = signalExitStatus
			}
			// we always have to wait for the main work to be finished, so
			// we exit only in this case
			os.Exit(exitStatus)
		}
	}
}

func registerSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	return signalChannel
}

func signalToExitStatus(signal os.Signal) int {
	switch signal {
	// kill -SIGINT XXXX or Ctrl+c
	case syscall.SIGINT:
		return 130
		// kill -SIGTERM XXXX, GoCD uses it to cancel tasks
	case syscall.SIGTERM:
		return 2
	default:
		return 99
	}
}
