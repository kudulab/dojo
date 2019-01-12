package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"
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
		PrintError(err.Error())
		os.Exit(1)
	}
	Log("debug", fmt.Sprintf("configFromCLI: %s", configFromCLI))
	Log("debug", fmt.Sprintf("configFromFile: %s", configFromFile))
	Log("debug", fmt.Sprintf("mergedConfig: %s", mergedConfig))
	Log("debug", fmt.Sprint("Config verified successfully"))
	return mergedConfig
}

func getEnvFilePath(runID string, test string) string {
	if test == "true" {
		return fmt.Sprintf("/tmp/test-dojo-environment-%s", runID)
	} else {
		return fmt.Sprintf("/tmp/dojo-environment-%s", runID)
	}
}

func handleRun(mergedConfig Config, runID string) int {
	exitStatus := 0
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	if currentUser.Username == "root" {
		Log("warn", "Current user is root, which is not recommended")
	}
	envFile := getEnvFilePath(runID, mergedConfig.Test)
	saveEnvToFile(envFile, mergedConfig.BlacklistVariables, mergedConfig.Dryrun)
	defer removeGeneratedFile(mergedConfig, envFile)
	interactiveShell := checkIfInteractive()
	Log("debug", fmt.Sprintf("Current shell is interactive: %v", interactiveShell))

	if mergedConfig.Driver == "docker"{
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
	} else {
		// driver: docker-compose
		dojoDCGeneratedFile := handleDCFiles(mergedConfig, envFile)
		defer removeGeneratedFile(mergedConfig, dojoDCGeneratedFile)
		cmd := constructDockerComposeCommandRun(mergedConfig, runID, dojoDCGeneratedFile, interactiveShell)
		cmdStop := constructDockerComposeCommandStop(mergedConfig, runID, dojoDCGeneratedFile)
		cmdRm := constructDockerComposeCommandRm(mergedConfig, runID, dojoDCGeneratedFile)
		Log("info", fmt.Sprintf("docker-compose run command will be:\n %v", cmd))
		Log("debug", fmt.Sprintf("docker-compose stop command will be:\n %v", cmdStop))
		Log("debug", fmt.Sprintf("docker-compose rm command will be:\n %v", cmdRm))
		expectedDockerNetwork := getExpDockerNetwork(runID)
		Log("debug", fmt.Sprintf("expected docker-compose network will be:\n %v", expectedDockerNetwork))
		if mergedConfig.Dryrun != "true" {
			exitStatus = RunShell(cmd)
			Log("debug", fmt.Sprintf("Exit status from run command: %v", exitStatus))
			Log("debug", "Stopping containers")
			RunShell(cmdStop)
			Log("debug", "Removing containers")
			RunShell(cmdRm)
			_, _, networkExists := RunShellGetOutput(fmt.Sprintf("docker network inspect %s", expectedDockerNetwork))
			if networkExists == 0 {
				Log("debug", fmt.Sprintf("Removing docker network: %s", expectedDockerNetwork))
				RunShell(fmt.Sprintf("docker network rm %s", expectedDockerNetwork))
			} else {
				Log("debug", fmt.Sprintf("Not removing docker network: %s, it does not exist", expectedDockerNetwork))
			}
		} else {
			Log("info", "Dryrun set, not running docker-compose")
		}
	}
	return exitStatus
}

func removeGeneratedFile(mergedConfig Config, filePath string) {
	if mergedConfig.Dryrun != "false" {
		Log("debug", fmt.Sprintf("Not removed generated file: %s, because dryrun is set", filePath))
		return
	}
	if mergedConfig.RemoveContainers != "false" {
		err := os.Remove(filePath)
		if err != nil {
			panic(err)
		}
		Log("debug", fmt.Sprintf("Removed generated file: %s", filePath))
		return
	} else {
		Log("debug", fmt.Sprintf("Not removed generated file: %s, because RemoveContainers is set", filePath))
		return
	}
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

func handleCleanupOnSignal(mergedConfig Config, runID string) {
	if mergedConfig.Action == "run" {
		if runID == "" {
			Log("debug", "No cleaning needed")
			return
		}
		if mergedConfig.RemoveContainers != "false" && mergedConfig.Dryrun != "true" {
			// add this sleep to let docker handle potential command we already sent to it
			time.Sleep(time.Second)

			stdout, _, _ := RunShellGetOutput(fmt.Sprintf("timeout 2 docker inspect -f '{{.State.Running}}' %s 2>&1", runID))
			if strings.Contains(stdout,"No such object"){
				Log("info", fmt.Sprintf("Not removing the container, it was not created: %s", runID))
			} else if strings.Contains(stdout,"false") {
				// Signal caught fast enough, docker container is not running, but
				// it is created. Let's remove it.
				Log("info", fmt.Sprintf("Removing created (but not started) container: %s", runID))
				stdout, _, exitStatus := RunShellGetOutput(fmt.Sprintf("docker rm %s 2>&1", runID))
				if exitStatus != 0 {
					PrintError(fmt.Sprintf("Exit status: %v, output: %s", exitStatus, stdout))
					os.Exit(exitStatus)
				}
			} else if strings.Contains(stdout,"true") {
				Log("info", fmt.Sprintf("Stopping running container: %s", runID))
				stdout, _, exitStatus := RunShellGetOutput(fmt.Sprintf("docker stop %s 2>&1", runID))
				if exitStatus != 0 {
					PrintError(fmt.Sprintf("Exit status: %v, output: %s", exitStatus, stdout))
					os.Exit(exitStatus)
				}
				// no need to remove the container, if it was started with "docker run --rm", it will be removed
			} else {
				// this is the case when docker is not installed
				Log("info", fmt.Sprintf("Not cleaning. Got output: %s", stdout))
			}
			envFile := getEnvFilePath(runID, mergedConfig.Test)
			removeGeneratedFile(mergedConfig, envFile)
		}
		Log("debug", "Cleanup finished")
	}
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
			PrintError(fmt.Sprintf("Caught signal: %s", signalCaught.String()))
			exitStatus := 1
			switch signalCaught {
			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				exitStatus = 130
				// kill -SIGTERM XXXX, GoCD uses it to cancel tasks
			case syscall.SIGTERM:
				exitStatus = 2
			}
			handleCleanupOnSignal(mergedConfig, runID)
			os.Exit(exitStatus)
		case done := <-doneChannel:
			Log("debug", "Done, normal way")
			os.Exit(done)
		}
	}
}
