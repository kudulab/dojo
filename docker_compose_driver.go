package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type DockerComposeDriver struct {
	ShellService ShellServiceInterface
	FileService  FileServiceInterface
}

func NewDockerComposeDriver(shellService ShellServiceInterface, fs FileServiceInterface) DockerComposeDriver {
	if shellService == nil {
		panic(errors.New("shellService was nil"))
	}
	if fs == nil {
		panic(errors.New("fs was nil"))
	}
	return DockerComposeDriver{
		ShellService: shellService,
		FileService: fs,
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
		Log("error", fmt.Sprintf("Docker-compose file %s is not correct: %s", mergedConfig.DockerComposeFile, err.Error()))
		return "", err
	}
	dojoDCFileContents := dc.generateDCFileContentsForRun(mergedConfig, version, envFile)
	dojoDCFileName := mergedConfig.DockerComposeFile + ".dojo"
	dc.FileService.WriteToFile(dojoDCFileName, dojoDCFileContents, "info")
	return dojoDCFileName, nil
}

func (dc DockerComposeDriver) generateDCFileContentsForRun(config Config, version float64, envFile string) string {
	if dc.FileService.GetFileUid(config.WorkDirOuter) == 0 {
		Log("warn", fmt.Sprintf("WorkDirOuter: %s is owned by root, which is not recommended", config.WorkDirOuter))
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

func (dc DockerComposeDriver) ConstructDockerComposeCommandPart1(config Config, projectName string, dojoGeneratedDCFile string) string {
	if projectName == "" {
		panic("projectName was not set")
	}
	if dojoGeneratedDCFile == "" {
		panic("dojoGeneratedDCFile was not set")
	}
	if config.DockerComposeFile == "" {
		panic("config.DockerComposeFile was not set")
	}
	cmd := fmt.Sprintf("docker-compose -f %s -f %s -p %s", config.DockerComposeFile, dojoGeneratedDCFile, projectName)
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandRun(config Config, projectName string, dojoGeneratedDCFile string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName, dojoGeneratedDCFile)
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
func (dc DockerComposeDriver) ConstructDockerComposeCommandStop(config Config, projectName string, dojoGeneratedDCFile string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName, dojoGeneratedDCFile)
	cmd += " stop"
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandRm(config Config, projectName string, dojoGeneratedDCFile string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName, dojoGeneratedDCFile)
	cmd += " rm -f"
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
		Log("warn", "Current user is root, which is not recommended")
	}
	envFile := getEnvFilePath(runID, mergedConfig.Test)
	saveEnvToFile(dc.FileService, envFile, mergedConfig.BlacklistVariables, envService.Variables())
	defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFile)

	dojoDCGeneratedFile, err := dc.handleDCFilesForRun(mergedConfig, envFile)
	if err != nil {
		return 1
	}
	defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, dojoDCGeneratedFile)

	cmd := dc.ConstructDockerComposeCommandRun(mergedConfig, runID, dojoDCGeneratedFile)
	Log("info", green(fmt.Sprintf("docker-compose run command will be:\n %v", cmd)))
	cmdStop := dc.ConstructDockerComposeCommandStop(mergedConfig, runID, dojoDCGeneratedFile)
	Log("debug", fmt.Sprintf("docker-compose stop command will be:\n %v", cmdStop))
	cmdRm := dc.ConstructDockerComposeCommandRm(mergedConfig, runID, dojoDCGeneratedFile)
	Log("debug", fmt.Sprintf("docker-compose rm command will be:\n %v", cmdRm))

	expectedDockerNetwork := dc.GetExpDockerNetwork(runID)
	Log("debug", fmt.Sprintf("expected docker-compose network will be:\n %v", expectedDockerNetwork))

	exitStatus := dc.ShellService.RunInteractive(cmd)
	Log("debug", fmt.Sprintf("Exit status from run command: %v", exitStatus))
	Log("debug", "Stopping containers")
	dc.ShellService.RunInteractive(cmdStop)
	Log("debug", "Removing containers")
	dc.ShellService.RunInteractive(cmdRm)
	_, _, networkExists := dc.ShellService.RunGetOutput(fmt.Sprintf("docker network inspect %s", expectedDockerNetwork))
	if networkExists == 0 {
		Log("debug", fmt.Sprintf("Removing docker network: %s", expectedDockerNetwork))
		stdout, stderr, es := dc.ShellService.RunGetOutput(fmt.Sprintf("docker network rm %s", expectedDockerNetwork))
		if es != 0 {
			Log("error", fmt.Sprintf("Error when removing docker network %s, exit status: %v\nstdout: %s\nstderr%s", expectedDockerNetwork, es, stdout, stderr))
			exitStatus = 1
		}
	} else {
		Log("debug", fmt.Sprintf("Not removing docker network: %s, it does not exist", expectedDockerNetwork))
	}
	return exitStatus
}

func (dc DockerComposeDriver) handleDCFilesForPull(mergedConfig Config) (string, error) {
	fileContents := dc.FileService.ReadDockerComposeFile(mergedConfig.DockerComposeFile)
	version, err := dc.verifyDCFile(fileContents, mergedConfig.DockerComposeFile)
	if err != nil {
		Log("error", fmt.Sprintf("Docker-compose file %s is not correct: %s", mergedConfig.DockerComposeFile, err.Error()))
		return "", err
	}
	dojoDCFileContents := dc.generateDCFileContentsForPull(mergedConfig, version)
	dojoDCFileName := mergedConfig.DockerComposeFile + ".dojo"
	dc.FileService.WriteToFile(dojoDCFileName, dojoDCFileContents, "info")
	return dojoDCFileName, nil
}

func (dc DockerComposeDriver) ConstructDockerComposeCommandPull(config Config, dojoGeneratedDCFile string) string {
	// projectName does not matter for docker-compose pull command
	cmd := dc.ConstructDockerComposeCommandPart1(config, "dojo", dojoGeneratedDCFile)
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
	Log("info", green(fmt.Sprintf("docker-compose pull command will be:\n %v", cmd)))
	exitStatus := dc.ShellService.RunInteractive(cmd)
	Log("debug", fmt.Sprintf("Exit status from pull command: %v", exitStatus))
	return exitStatus
}

func (d DockerComposeDriver) HandleSignal(mergedConfig Config, runID string) int {
	Log("debug", "Additional cleanup not needed")
	return 0
}