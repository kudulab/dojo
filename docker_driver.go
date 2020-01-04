package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type DockerDriver struct {
	ShellService ShellServiceInterface
	FileService  FileServiceInterface
	Logger *Logger
}

func NewDockerDriver(shellService ShellServiceInterface, fs FileServiceInterface, logger *Logger) DockerDriver {
	if shellService == nil {
		panic(errors.New("shellService was nil"))
	}
	if fs == nil {
		panic(errors.New("fs was nil"))
	}
	if logger == nil {
		panic(errors.New("logger was nil"))
	}
	return DockerDriver{
		ShellService: shellService,
		FileService: fs,
		Logger: logger,
	}
}

func (d DockerDriver) ConstructDockerRunCmd(config Config, envFilePath string, envFileMultiLine string, containerName string) string {
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
	cmd += fmt.Sprintf(" -v %s:%s -v %s:/dojo/identity:ro -v %s:/etc/dojo.d/variables/00-multiline-vars.sh",
			config.WorkDirOuter, config.WorkDirInner, config.IdentityDirOuter, envFileMultiLine)
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
	warnGeneral(d.FileService, mergedConfig, envService, d.Logger)
	envFile, envFileMultiLine := getEnvFilePaths(runID, mergedConfig.Test)
	saveEnvToFile(d.FileService, envFile, envFileMultiLine, mergedConfig.BlacklistVariables, envService.GetVariables())

	cmd := d.ConstructDockerRunCmd(mergedConfig, envFile, envFileMultiLine, runID)
	d.Logger.Log("info", green(fmt.Sprintf("docker command will be:\n %v", cmd)))

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
	exitStatus, _ := d.ShellService.RunInteractive(cmd, true)
	d.Logger.Log("debug", fmt.Sprintf("Exit status from run command: %v", exitStatus))
	return exitStatus

	// do not clean now, container may be being stopped in other goroutines
}

func (d DockerDriver) HandlePull(mergedConfig Config) int {
	cmd := fmt.Sprintf("docker pull %s", mergedConfig.DockerImage)
	d.Logger.Log("info", green(fmt.Sprintf("docker pull command will be:\n %v", cmd)))
	exitStatus, _ := d.ShellService.RunInteractive(cmd, false)
	d.Logger.Log("debug", fmt.Sprintf("Exit status from pull command: %v", exitStatus))
	return exitStatus
}

// Stop the container if it is not removed.
func (d DockerDriver) HandleSignal(mergedConfig Config, runID string) int {
	d.Logger.Log("info", "Stopping on signal")
	containerInfo, err := getContainerInfo(d.ShellService, runID)
	if err != nil {
		d.Logger.Log("info", fmt.Sprintf("Not cleaning. Unexpected error.\n%s", err))
		panic(err)
	}
	if !containerInfo.Exists {
		d.Logger.Log("info", "Container already removed or not created at all")
		pid := d.ShellService.GetProcessPid([]string{"docker run", runID})
		if pid != 0 {
			proc, err := os.FindProcess(pid)
			if err != nil {
				d.Logger.Log("debug", fmt.Sprintf("Nothing to do, docker run process not found: %s", err.Error()))
				panic(err)
			}
			d.Logger.Log("info", fmt.Sprintf("Sending SIGINT to the docker run process: %s, to stop the docker pull", strconv.Itoa(pid)))
			err = proc.Signal(os.Interrupt)
			if err != nil {
				panic(err)
			}
		}
		return 0
	}
	cmd := fmt.Sprintf("docker stop %s", runID)
	d.Logger.Log("info", fmt.Sprintf("Stopping container with command: \n%v", cmd))
	exitStatus, _ := d.ShellService.RunInteractive(cmd, false)
	d.Logger.Log("debug", fmt.Sprintf("Exit status from command: %s, %v", cmd, exitStatus))
	d.Logger.Log("info", "Stopping on signal finished")
	return exitStatus
}

// Kill the container if it is not removed.
func (d DockerDriver) HandleMultipleSignal(mergedConfig Config, runID string) int {
	d.Logger.Log("info", "Stopping on multiple signals")
	containerInfo, err := getContainerInfo(d.ShellService, runID)
	if err != nil {
		d.Logger.Log("info", fmt.Sprintf("Not cleaning. Unexpected error.\n%s", err))
		panic(err)
	}
	if !containerInfo.Exists {
		d.Logger.Log("info", "Container already removed or not created at all, will not react on this signal")
		return 0
	}
	cmd := fmt.Sprintf("docker kill %s", runID)
	d.Logger.Log("info", fmt.Sprintf("Stopping container with command: \n%v", cmd))
	stdout, stderr, exitStatus, _ := d.ShellService.RunGetOutput(cmd, true)
	if exitStatus != 0 {
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
		d.Logger.Log("debug", cmdInfo)
	} else {
		d.Logger.Log("debug", "docker kill was successful")
	}
	d.Logger.Log("info", "Stopping on multiple signals finished")
	return exitStatus
}

func (d DockerDriver) CleanAfterRun(mergedConfig Config, runID string) int {
	if mergedConfig.RemoveContainers == "true" {
		d.Logger.Log("debug", "Cleaning, because RemoveContainers is set to true")
		envFile, envFileMultiLine := getEnvFilePaths(runID, mergedConfig.Test)
		d.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFile)
		d.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFileMultiLine)

		// no need to remove the container, if it was started with "docker run --rm", it was already removed

		return 0
	} else {
		d.Logger.Log("debug", "Not cleaning, because RemoveContainers is not set to true")
		return 0
	}
}
