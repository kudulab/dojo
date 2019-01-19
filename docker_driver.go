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
		cmd += fmt.Sprintf(" %s", config.RunCommand)
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

// If a container's process preserves signals, the container should be stopped relatively fast
// (usually after several seconds)
func (d DockerDriver) waitForContainerRemoval(cmd string) bool {
	timeout := 12
	Log("info", fmt.Sprintf("Waiting max %vs for container removal", timeout))
	for i:=0; i<timeout; i++ {
		stdout, _, _ := d.ShellService.RunGetOutput(cmd)
		if strings.Contains(stdout, "No such object") {
			Log("info", "Container removed by docker")
			return true
		}
		Log("info", fmt.Sprintf("Trial: %v", i))
		time.Sleep(time.Second)
	}
	Log("info", fmt.Sprintf("Container not removed after %vs", timeout))
	return false
}

// HandleSignal follows such a procedure:
// * if container does not exist, just exit
// * otherwise, let's wait for a container to be removed by docker (if it preserves signals, it will be removed)
//   * if container was removed - ensure env file is removed and exit
//   * if container was not removed - let's stop it (if it was run with docker run --rm, then docker will remove it)
//
// Any default timeout is here slightly bigger than 10s, because it takes 10s to docker stop containers which
// do not preserve signals.
func (d DockerDriver) HandleSignal(mergedConfig Config, runID string) int {
	if mergedConfig.RemoveContainers != "false" {
		Log("debug", "Cleaning on signal")

		cmd := fmt.Sprintf("timeout 2 docker inspect -f '{{.State.Running}}' %s 2>&1", runID)

		stdout, stderr, exitStatus := d.ShellService.RunGetOutput(cmd)
		if strings.Contains(stdout, "No such object") {
			Log("info", fmt.Sprintf("Cleaning not needed, the container was not created: %s", runID))
		} else if exitStatus != 0 {
			// unexpected error case
			cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
			Log("info", fmt.Sprintf("Not cleaning. Unexpected non-0 exit status.\n%s", cmdInfo))
		} else {
			// Container is either:
			// * created and not running
			// * or running
			// It still may be removed by docker.
			containerRemoved := d.waitForContainerRemoval(cmd)
			if !containerRemoved {
				Log("info", fmt.Sprintf("Stopping container: %s", runID))
				stdout, stderr, exitStatus := d.ShellService.RunGetOutput(fmt.Sprintf("docker stop %s 2>&1", runID))
				if exitStatus != 0 {
					cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
					Log("error", cmdInfo)
					return exitStatus
				}
				// no need to remove the container, if it was started with "docker run --rm", it will be removed
			}
		}

		envFile := getEnvFilePath(runID, mergedConfig.Test)
		// if container was removed by docker, the env file may be already removed too,
		// so let's ignore error on removal here
		d.FileService.RemoveGeneratedFileIgnoreError(mergedConfig.RemoveContainers, envFile, true)
		Log("info", "Cleanup finished")
	}
	return 0
}

func (d DockerDriver) HandleMultipleSignal(mergedConfig Config, runID string) int {
	if mergedConfig.RemoveContainers != "false" {
		Log("debug", "Multiple signals caught, running docker kill and immediate exit")
		if !d.FileService.FileExists(getEnvFilePath(runID, mergedConfig.Test)) {
			Log("debug", "Environment file does not exist, either containers not created or already cleaned")
			return 0
		}
		cmd := fmt.Sprintf("docker kill %s", runID)
		stdout, stderr, exitStatus := d.ShellService.RunGetOutput(fmt.Sprintf("docker kill %s", runID))
		if exitStatus != 0 {
			cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
			Log("debug", cmdInfo)
		} else {
			Log("debug", "docker kill was successful")
		}
		return exitStatus
	}
	return 0
}

