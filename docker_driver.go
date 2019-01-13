package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type DockerDriver struct {
	ShellService ShellServiceInterface
	FileService  FileServiceInterface
}

func NewDockerDriver(shellService ShellServiceInterface, fs FileServiceInterface) DockerDriver {
	if shellService == nil {
		panic(errors.New("shellService was nil"))
	}
	if fs == nil {
		panic(errors.New("fs was nil"))
	}
	return DockerDriver{
		ShellService: shellService,
		FileService: fs,
	}
}

func (d DockerDriver) ConstructDockerRunCmd(config Config, envFilePath string, containerName string) string {
	if envFilePath == "" {
		panic("envFilePath was not set")
	}
	if containerName == "" {
		panic("containerName was not set")
	}
	cmd := "docker run"
	if config.RemoveContainers == "true" {
		cmd += " --rm"
	}
	if d.FileService.GetFileUid(config.WorkDirOuter) == 0 {
		Log("warn", fmt.Sprintf("WorkDirOuter: %s is owned by root, which is not recommended", config.WorkDirOuter))
	}
	cmd += fmt.Sprintf(" -v %s:%s -v %s:/dojo/identity:ro", config.WorkDirOuter, config.WorkDirInner, config.IdentityDirOuter)
	cmd += fmt.Sprintf(" --env-file=%s", envFilePath)
	if os.Getenv("DISPLAY") != "" {
		// DISPLAY is set, enable running in graphical mode (opinionated)
		cmd += " -v /tmp/.X11-unix:/tmp/.X11-unix"
	}
	if config.DockerOptions != "" {
		cmd += fmt.Sprintf(" %s", config.DockerOptions)
	}
	shellIsInteractive := d.ShellService.CheckIfInteractive()
	if config.Interactive == "true" {
		cmd += " -ti"
	} else if config.Interactive == "false"  {
		// nothing
	} else if shellIsInteractive {
		cmd += " -ti"
	}
	cmd += fmt.Sprintf(" --name=%s", containerName)
	cmd += fmt.Sprintf(" %s", config.DockerImage)
	if config.RunCommand != "" {
		if strings.Contains(config.RunCommand, "\"") {
			// command contains quotes or is wrapped with quotes, do not wrap it again
			cmd += fmt.Sprintf(" %s", config.RunCommand)
		} else {
			// wrap command with quotes
			cmd += fmt.Sprintf(" \"%s\"", config.RunCommand)
		}
	}
	return cmd
}

func (d DockerDriver) HandleRun(mergedConfig Config, runID string, envService EnvServiceInterface) int {
	if envService.IsCurrentUserRoot() {
		Log("warn", "Current user is root, which is not recommended")
	}
	envFile := getEnvFilePath(runID, mergedConfig.Test)
	saveEnvToFile(d.FileService, envFile, mergedConfig.BlacklistVariables, envService.Variables())
	defer d.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFile)

	cmd := d.ConstructDockerRunCmd(mergedConfig, envFile, runID)
	Log("info", green(fmt.Sprintf("docker command will be:\n %v", cmd)))

	if mergedConfig.RemoveContainers != "true" {
		// Removing docker container is impractical without additional steps here. We'd have to
		// parse output of dojo in order to get container name. Thus, save the container name to a file.
		currentDirectory := d.FileService.GetCurrentDir()
		rcFile := fmt.Sprintf("%s/dojorc.txt", currentDirectory)
		d.FileService.RemoveFile(rcFile, true)
		d.FileService.WriteToFile(rcFile, runID, "info")

		rcFile2 := fmt.Sprintf("%s/dojorc", currentDirectory)
		d.FileService.RemoveFile(rcFile2, true)
		d.FileService.WriteToFile(rcFile2, fmt.Sprintf("DOJO_RUN_ID=%s",runID), "info")
	}
	exitStatus := d.ShellService.RunInteractive(cmd)
	Log("debug", fmt.Sprintf("Exit status: %v", exitStatus))
	return exitStatus
}

func (d DockerDriver) HandlePull(mergedConfig Config) int {
	cmd := fmt.Sprintf("docker pull %s", mergedConfig.DockerImage)
	Log("info", green(fmt.Sprintf("docker pull command will be:\n %v", cmd)))
	exitStatus := d.ShellService.RunInteractive(cmd)
	Log("debug", fmt.Sprintf("Exit status from pull command: %v", exitStatus))
	return exitStatus
}

func (d DockerDriver) HandleSignal(mergedConfig Config, runID string) int {
	if mergedConfig.RemoveContainers != "false" {
		// add this sleep to let docker handle potential command we already sent to it
		time.Sleep(time.Second)

		stdout, _, _ := d.ShellService.RunGetOutput(fmt.Sprintf("timeout 2 docker inspect -f '{{.State.Running}}' %s 2>&1", runID))
		if strings.Contains(stdout, "No such object") {
			Log("info", fmt.Sprintf("Not removing the container, it was not created: %s", runID))
		} else if strings.Contains(stdout, "false") {
			// Signal caught fast enough, docker container is not running, but
			// it is created. Let's remove it.
			Log("info", fmt.Sprintf("Removing created (but not started) container: %s", runID))
			stdout, _, exitStatus := d.ShellService.RunGetOutput(fmt.Sprintf("docker rm %s 2>&1", runID))
			if exitStatus != 0 {
				Log("error", fmt.Sprintf("Exit status: %v, output: %s", exitStatus, stdout))
				return exitStatus
			}
		} else if strings.Contains(stdout, "true") {
			Log("info", fmt.Sprintf("Stopping running container: %s", runID))
			stdout, _, exitStatus := d.ShellService.RunGetOutput(fmt.Sprintf("docker stop %s 2>&1", runID))
			if exitStatus != 0 {
				Log("error", fmt.Sprintf("Exit status: %v, output: %s", exitStatus, stdout))
				return exitStatus
			}
			// no need to remove the container, if it was started with "docker run --rm", it will be removed
		} else {
			// this is the case when docker is not installed
			Log("info", fmt.Sprintf("Not cleaning. Got output: %s", stdout))
		}
		envFile := getEnvFilePath(runID, mergedConfig.Test)
		d.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFile)
	}
	Log("debug", "Cleanup finished")
	return 0
}

