package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type DockerComposeDriver struct {
	ShellService ShellServiceInterface
	FileService  FileServiceInterface
	Logger *Logger
}

func NewDockerComposeDriver(shellService ShellServiceInterface, fs FileServiceInterface, logger *Logger) DockerComposeDriver {
	if shellService == nil {
		panic(errors.New("shellService was nil"))
	}
	if fs == nil {
		panic(errors.New("fs was nil"))
	}
	if logger == nil {
		panic(errors.New("logger was nil"))
	}
	return DockerComposeDriver{
		ShellService: shellService,
		FileService: fs,
		Logger: logger,
	}
}

func (dc DockerComposeDriver) parseDCFileVersion(contents string) (float64, error) {
	firstLine := strings.Split(contents, "\n")[0]
	if !strings.HasPrefix(firstLine,"version") {
		return 0, fmt.Errorf("First line of docker-compose file did not start with: version")
	}
	versionQuoted := strings.Split(firstLine, ":")[1]
	versionQuoted = strings.Trim(versionQuoted, " ")
	versionQuoted = strings.TrimSuffix(versionQuoted,"\n")
	versionQuoted = strings.Trim(versionQuoted, "\"")
	versionQuoted = strings.Trim(versionQuoted, "'")
	version, err := strconv.ParseFloat(versionQuoted, 64)
	if err != nil {
		return 0, err
	}
	return version,nil
}
func (dc DockerComposeDriver) verifyDCFile(fileContents string, filePath string) (float64, error) {
	version, err := dc.parseDCFileVersion(fileContents)
	if err != nil {
		return 0,err
	}
	if version < 2 || version >= 3 {
		return 0,fmt.Errorf("docker-compose file: %s should contain version number >=2 and <3, current version: %v", filePath, version)
	}
	requiredStr := "default:"
	if ! strings.Contains(fileContents, requiredStr) {
		return 0,fmt.Errorf("docker-compose file: %s does not contain: %s. Please add a default service", filePath, requiredStr)
	}
	return version, nil
}

func (dc DockerComposeDriver) handleDCFilesForRun(mergedConfig Config, envFile string) (string, error) {
	fileContents := dc.FileService.ReadDockerComposeFile(mergedConfig.DockerComposeFile)
	version, err := dc.verifyDCFile(fileContents, mergedConfig.DockerComposeFile)
	if err != nil {
		dc.Logger.Log("error", fmt.Sprintf("Docker-compose file %s is not correct: %s", mergedConfig.DockerComposeFile, err.Error()))
		return "", err
	}
	dojoDCFileContents := dc.generateDCFileContentsForRun(mergedConfig, version, envFile)
	dojoDCFileName := dc.getDCGeneratedFilePath(mergedConfig.DockerComposeFile)
	dc.FileService.WriteToFile(dojoDCFileName, dojoDCFileContents, "info")
	return dojoDCFileName, nil
}

func (dc DockerComposeDriver) getDCGeneratedFilePath(dcfilePath string) string {
	return dcfilePath + ".dojo"
}

func (dc DockerComposeDriver) generateDCFileContentsForRun(config Config, version float64, envFile string) string {
	if dc.FileService.GetFileUid(config.WorkDirOuter) == 0 {
		dc.Logger.Log("warn", fmt.Sprintf("WorkDirOuter: %s is owned by root, which is not recommended", config.WorkDirOuter))
	}
	contents := fmt.Sprintf(
		`version: '%v'
services:
  default:
    image: %s
    volumes:
      - %s:%s:ro
      - %s:%s
`, version, config.DockerImage, config.IdentityDirOuter, "/dojo/identity", config.WorkDirOuter, config.WorkDirInner)
	if os.Getenv("DISPLAY") != "" {
		// DISPLAY is set, enable running in graphical mode (opinionated)
		contents += "      - /tmp/.X11-unix:/tmp/.X11-unix\n"
	}
	contents += fmt.Sprintf(`    env_file:
      - %s
`, envFile)
	return contents
}

func (dc DockerComposeDriver) generateDCFileContentsForPull(config Config, version float64) string {
	contents := fmt.Sprintf(
		`version: '%v'
services:
  default:
    image: %s
`, version, config.DockerImage)
	return contents
}

func (dc DockerComposeDriver) ConstructDockerComposeCommandPart1(config Config, projectName string) string {
	if projectName == "" {
		panic("projectName was not set")
	}
	if config.DockerComposeFile == "" {
		panic("config.DockerComposeFile was not set")
	}
	dcGenFile := dc.getDCGeneratedFilePath(config.DockerComposeFile)
	cmd := fmt.Sprintf("docker-compose -f %s -f %s -p %s", config.DockerComposeFile, dcGenFile, projectName)
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandRun(config Config, projectName string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName)
	cmd += " run --rm"
	shellIsInteractive := dc.ShellService.CheckIfInteractive()
	if !shellIsInteractive && config.RunCommand == "" {
		// TODO: Support this case. Maybe use docker-compose up instead of docker-compose run? #17186
		panic("Using driver: docker-compose with empty RunCommand when shell is not interactive is unsupported. It would hang the terminal")
	}
	if config.Interactive == "false" {
		cmd += " -T"
	} else if config.Interactive == "true"  {
		// nothing
	} else if !shellIsInteractive {
		cmd += " -T"
	}
	if config.DockerComposeOptions != "" {
		cmd += fmt.Sprintf(" %s", config.DockerComposeOptions)
	}
	cmd += " default"

	if config.RunCommand != "" {
		cmd += fmt.Sprintf(" %s", config.RunCommand)
	}
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandStop(config Config, projectName string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName)
	cmd += " stop"
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandRm(config Config, projectName string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName)
	cmd += " rm -f"
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandPs(config Config, projectName string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName)
	cmd += " ps -q"
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandDown(config Config, projectName string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName)
	cmd += " down"
	return cmd
}
func (dc DockerComposeDriver) GetExpDockerNetwork(runID string) string {
	// remove dashes and underscores and add "_default" (we always demand docker container named: "default")
	runIDPrim := strings.Replace(runID, "/","", -1)
	runIDPrim = strings.Replace(runIDPrim, "_","", -1)
	runIDPrim = strings.Replace(runIDPrim, "-","", -1)
	network := fmt.Sprintf("%s_default", runIDPrim)
	return network
}
func (dc DockerComposeDriver) HandleRun(mergedConfig Config, runID string, envService EnvServiceInterface) int {
	if envService.IsCurrentUserRoot() {
		dc.Logger.Log("warn", "Current user is root, which is not recommended")
	}
	envFile := getEnvFilePath(runID, mergedConfig.Test)
	saveEnvToFile(dc.FileService, envFile, mergedConfig.BlacklistVariables, envService.Variables())
	// this file might have already been removed by one of HandleSignal() functions
	// so let's ignore error on removal here
	defer dc.FileService.RemoveGeneratedFileIgnoreError(mergedConfig.RemoveContainers, envFile, true)

	dojoDCGeneratedFile, err := dc.handleDCFilesForRun(mergedConfig, envFile)
	if err != nil {
		return 1
	}
	// this file might have already been removed by one of HandleSignal() functions
	// so let's ignore error on removal here
	defer dc.FileService.RemoveGeneratedFileIgnoreError(mergedConfig.RemoveContainers, dojoDCGeneratedFile, true)

	cmd := dc.ConstructDockerComposeCommandRun(mergedConfig, runID)
	dc.Logger.Log("info", green(fmt.Sprintf("docker-compose run command will be:\n %v", cmd)))
	cmdStop := dc.ConstructDockerComposeCommandStop(mergedConfig, runID)
	dc.Logger.Log("debug", fmt.Sprintf("docker-compose stop command will be:\n %v", cmdStop))
	cmdRm := dc.ConstructDockerComposeCommandRm(mergedConfig, runID)
	dc.Logger.Log("debug", fmt.Sprintf("docker-compose rm command will be:\n %v", cmdRm))

	expectedDockerNetwork := dc.GetExpDockerNetwork(runID)
	dc.Logger.Log("debug", fmt.Sprintf("expected docker-compose network will be:\n %v", expectedDockerNetwork))

	exitStatus := dc.ShellService.RunInteractive(cmd)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from run command: %v", exitStatus))

	dc.Logger.Log("debug", "Stopping containers")
	exitStatusStop := dc.ShellService.RunInteractive(cmdStop)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from stop command: %v", exitStatusStop))

	dc.Logger.Log("debug", "Removing containers")
	exitStatusRm := dc.ShellService.RunInteractive(cmdRm)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from rm command: %v", exitStatusRm))

	_, _, networkExists := dc.ShellService.RunGetOutput(fmt.Sprintf("docker network inspect %s", expectedDockerNetwork))
	if networkExists == 0 {
		dc.Logger.Log("debug", fmt.Sprintf("Removing docker network: %s", expectedDockerNetwork))
		cmdDockerNetRm := fmt.Sprintf("docker network rm %s", expectedDockerNetwork)
		stdout, stderr, es := dc.ShellService.RunGetOutput(cmdDockerNetRm)
		if es != 0 {
			cmdInfo := cmdInfoToString(cmdDockerNetRm, stdout, stderr, es)
			dc.Logger.Log("error", cmdInfo)
			exitStatus = 1
		}
	} else {
		dc.Logger.Log("debug", fmt.Sprintf("Not removing docker network: %s, it does not exist", expectedDockerNetwork))
	}
	if exitStatusStop != 0 {
		exitStatus = exitStatusStop
	}
	if exitStatusRm != 0 {
		exitStatus = exitStatusRm
	}
	return exitStatus
}

func (dc DockerComposeDriver) handleDCFilesForPull(mergedConfig Config) (string, error) {
	fileContents := dc.FileService.ReadDockerComposeFile(mergedConfig.DockerComposeFile)
	version, err := dc.verifyDCFile(fileContents, mergedConfig.DockerComposeFile)
	if err != nil {
		dc.Logger.Log("error", fmt.Sprintf("Docker-compose file %s is not correct: %s", mergedConfig.DockerComposeFile, err.Error()))
		return "", err
	}
	dojoDCFileContents := dc.generateDCFileContentsForPull(mergedConfig, version)
	dojoDCFileName := mergedConfig.DockerComposeFile + ".dojo"
	dc.FileService.WriteToFile(dojoDCFileName, dojoDCFileContents, "info")
	return dojoDCFileName, nil
}

func (dc DockerComposeDriver) ConstructDockerComposeCommandPull(config Config, dojoGeneratedDCFile string) string {
	// projectName does not matter for docker-compose pull command
	cmd := dc.ConstructDockerComposeCommandPart1(config, "dojo")
	cmd += " pull"
	return cmd
}

func (dc DockerComposeDriver) HandlePull(mergedConfig Config) int {
	dojoDCGeneratedFile, err := dc.handleDCFilesForPull(mergedConfig)
	if err != nil {
		return 1
	}
	defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, dojoDCGeneratedFile)

	cmd := dc.ConstructDockerComposeCommandPull(mergedConfig, dojoDCGeneratedFile)
	dc.Logger.Log("info", green(fmt.Sprintf("docker-compose pull command will be:\n %v", cmd)))
	exitStatus := dc.ShellService.RunInteractive(cmd)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from pull command: %v", exitStatus))
	return exitStatus
}

func (dc DockerComposeDriver) checkContainersRemoved(cmd string) bool {
	stdout, _, es := dc.ShellService.RunGetOutput(cmd)
	if stdout == "" && es == 0 {
		dc.Logger.Log("info", "Containers removed by docker-compose")
		return true
	}
	dc.Logger.Log("info", "Containers not removed")
	return false
}

func (dc DockerComposeDriver) HandleSignal(mergedConfig Config, runID string) int {
	if mergedConfig.RemoveContainers != "false" {
		dc.Logger.Log("debug", "Cleaning on signal")
		fileRemoved := dc.waitForEnvFileRemoval(getEnvFilePath(runID, mergedConfig.Test))
		if fileRemoved {
			// cleaned up without additional help
			return 0
		}

		cmd := dc.ConstructDockerComposeCommandPs(mergedConfig, runID)

		stdout, stderr, exitStatus := dc.ShellService.RunGetOutput(cmd)
		if stdout == "" && exitStatus == 0 {
			dc.Logger.Log("info", fmt.Sprintf("Cleaning not needed, the containers were not created: %s", runID))
		} else if exitStatus != 0 {
			// unexpected error case
			dc.Logger.Log("info", "Not cleaning")
			dc.Logger.Log("info", cmdInfoToString(cmd, stdout, stderr, exitStatus))
		} else {
			// Containers are either:
			// * created and not running
			// * or running
			// They still may be removed by docker-compose.
			containersRemoved := dc.checkContainersRemoved(cmd)
			if !containersRemoved {
				dc.Logger.Log("info", fmt.Sprintf("Stopping docker-compose project: %s", runID))
				// docker-compose stop does not wait until containers are stopped, thus use this instead:
				// do not use "rm --force --stop", because "down" also removes networks
				stopCmd := dc.ConstructDockerComposeCommandDown(mergedConfig, runID)
				stdout, stderr, exitStatus := dc.ShellService.RunGetOutput(stopCmd)
				if exitStatus != 0 {
					dc.Logger.Log("error", cmdInfoToString(stopCmd, stdout, stderr, exitStatus))
					return exitStatus
				}
			}
		}

		envFile := getEnvFilePath(runID, mergedConfig.Test)
		// if containers were removed by docker-compose, the env file may be already removed too,
		// so let's ignore error on removal here
		dc.FileService.RemoveGeneratedFileIgnoreError(mergedConfig.RemoveContainers, envFile, true)
		dc.Logger.Log("info", "Cleanup finished")
	}
	return 0
}

func (dc DockerComposeDriver) waitForEnvFileRemoval(envFile string) bool {
	// 12 is too little
	timeout := 20
	dc.Logger.Log("info", fmt.Sprintf("Waiting max %vs for environment file to be removed", timeout))
	for i:=0; i<timeout; i++ {
		if !dc.FileService.FileExists(envFile) {
			dc.Logger.Log("info", "Containers removed by docker-compose")
			return true
		}
		dc.Logger.Log("info", fmt.Sprintf("Trial: %v", i))
		time.Sleep(time.Second)
	}
	dc.Logger.Log("info", fmt.Sprintf("Environment file not removed after %vs", timeout))
	return false
}

func (dc DockerComposeDriver) HandleMultipleSignal(mergedConfig Config, runID string) int {
	dc.Logger.Log("debug", "Multiple signals caught")
	// On two Ctrl+C signals docker-compose reacts and stops and removes (kills?) the docker containers.
	// We shall not perform any additional cleanup
	// So let's just wait for the environment file to be removed
	fileRemoved := dc.waitForEnvFileRemoval(getEnvFilePath(runID, mergedConfig.Test))
	if fileRemoved {
		return 0
	}
	return 1
}