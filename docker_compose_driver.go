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
	// This channel is closed when the stop action is started
	Stopping 	 chan bool
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
	stopping := make(chan bool)
	return DockerComposeDriver{
		ShellService: shellService,
		FileService: fs,
		Logger: logger,
		Stopping: stopping,
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


func (dc DockerComposeDriver) getDCGeneratedFilePath(dcfilePath string) string {
	return dcfilePath + ".dojo"
}

func (dc DockerComposeDriver) generateDCFileContentsWithEnv(expContainers []string,	config Config, envFile string,
	envFileMultiLine string) string {
	contents := fmt.Sprintf(
		`    volumes:
      - %s:%s:ro
      - %s:%s
      - %s:/etc/dojo.d/variables/00-multiline-vars.sh
`, config.IdentityDirOuter, "/dojo/identity", config.WorkDirOuter, config.WorkDirInner, envFileMultiLine)
	if os.Getenv("DISPLAY") != "" {
		// DISPLAY is set, enable running in graphical mode (opinionated)
		contents += "      - /tmp/.X11-unix:/tmp/.X11-unix\n"
	}
	contents += fmt.Sprintf(`    env_file:
      - %s
`, envFile)

	if config.PreserveEnvironmentToAllContainers == "true" {
		// set the env_file for each container
		for _, name := range expContainers {
			if name == "default" {
				// handled above
				continue
			}
			contents += fmt.Sprintf(`  %s:
    env_file:
      - %s
    volumes:
      - %s:/etc/dojo.d/variables/00-multiline-vars.sh
`, name, envFile, envFileMultiLine)
		}
	}
	return contents
}

func (dc DockerComposeDriver) generateInitialDCFile(config Config, version float64) string {
	contents := fmt.Sprintf(
		`version: '%v'
services:
  default:
    image: %s`, version, config.DockerImage)
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
func (dc DockerComposeDriver) ConstructDockerComposeCommandPs(config Config, projectName string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName)
	cmd += " ps"
	return cmd
}
func (dc DockerComposeDriver) ConstructDockerComposeCommandDown(config Config, projectName string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(config, projectName)
	cmd += " down"
	return cmd
}

// A method to check whether or not a channel is closed if you can make sure no values were ever sent to the channel.
// https://go101.org/article/channel-closing.html
func isChannelClosed(ch <-chan bool) bool {
	select {
	case <- ch:
		return true
	default:
	}
	return false
}

func (dc DockerComposeDriver) checkAllContainersRunning(containerIDs []string) (bool) {
	for _, id := range containerIDs {
		running := dc.checkContainerIsRunning(id)
		if !running {
			return false
		}
	}
	return true
}

func (dc DockerComposeDriver) checkContainerIsRunning(containerID string) (bool) {
	containerInfo, err := getContainerInfo(dc.ShellService, containerID)
	if err != nil {
		panic(err)
	}
	if !containerInfo.Exists {
		dc.Logger.Log("debug",
			fmt.Sprintf("Expected container: %s to be running, but it does not exist", containerID))
		return false
	}
	if containerInfo.Status != "running" {
		dc.Logger.Log("debug",
			fmt.Sprintf("Expected container: %s to be running, but was: %s", containerID, containerInfo.Status))
		return false
	}
	return true
}

// returns expected containers names as written in a docker-compose file, e.g. default, abc
func (dc DockerComposeDriver) getExpectedContainers(mergedConfig Config, runID string) []string {
	cmd := dc.ConstructDockerComposeCommandPart1(mergedConfig, runID)
	cmd += " config --services"
	stdout, stderr, es, _ := dc.ShellService.RunGetOutput(cmd, true)
	if es != 0 || stderr != "" || stdout == "" {
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, es)
		panic(fmt.Errorf("Unexpected error: %s", cmdInfo))
	}

	stdout = strings.TrimSuffix(stdout, "\n")
	lines := strings.Split(stdout, "\n")
	return lines
}

// Run the docker-compose ps command until it returns some output - containers IDs and
// then run docker inspect on each container. Return if all the containers are running.
//
// When docker-compose containers start, docker-compose ps may return not all the containers, because some of them
// may be not created yet. Thus, we have to know the number of containers specified in docker-compose config file - expContainersCount.
func (dc DockerComposeDriver) waitForContainersToBeRunning(mergedConfig Config, runID string, expContainersCount int) ([]string) {
	dc.Logger.Log("debug", fmt.Sprintf("Start waiting for containers to be initally running, %s", runID))

	for {
		if isChannelClosed(dc.Stopping) {
			dc.Logger.Log("debug", fmt.Sprintf("Not waiting anymore for containers %s", runID))
			return []string{}
		}
		containersNames := dc.getDCContainersNames(mergedConfig, runID)
		if len(containersNames) == 0 {
			dc.Logger.Log("debug", fmt.Sprintf("Containers not yet created: %s", runID))
			time.Sleep(time.Second)
			continue
		} else if len(containersNames) != expContainersCount {
			dc.Logger.Log("debug", fmt.Sprintf(
				"Not all the containers created: %s. Want: %v, have: %v", runID, expContainersCount, len(containersNames)))
			time.Sleep(time.Second)
			continue
		} else {
			dc.Logger.Log("debug",
				fmt.Sprintf("Containers created. Waiting for them to be initially running: %v", containersNames))
			allRunning := dc.checkAllContainersRunning(containersNames)
			if allRunning {
				dc.Logger.Log("debug", "All containers are running")
				return containersNames
			} else {
				time.Sleep(time.Second)
				continue
			}
		}
	}
}

func (dc DockerComposeDriver) watchContainers(mergedConfig Config, runID string, expContainersCount int) {
	if mergedConfig.ExitBehavior == "ignore" {
		return
	}
	dc.Logger.Log("debug", fmt.Sprintf(
		"Start watching docker-compose containers %s in a forever loop, exitBehavior is: %s",
		runID, mergedConfig.ExitBehavior))

	names := dc.waitForContainersToBeRunning(mergedConfig, runID, expContainersCount)
	for {
		if isChannelClosed(dc.Stopping) {
			dc.Logger.Log("debug", fmt.Sprintf("Stop watching docker-compose containers %s", runID))
			return
		}

		for _, name := range names {
			if isChannelClosed(dc.Stopping) {
				dc.Logger.Log("debug", fmt.Sprintf("Stop watching docker-compose containers %s", runID))
				return
			}
			running := dc.checkContainerIsRunning(name)
			if !running {
				if mergedConfig.ExitBehavior == "restart" {
					dc.Logger.Log("info", fmt.Sprintf("Container: %s stopped by itself. Starting...", name))
					cmd := fmt.Sprintf("docker start %s", name)
					stdout, stderr, exitStatus, _ := dc.ShellService.RunGetOutput(cmd, false)
					ci := cmdInfoToString(cmd, stdout, stderr, exitStatus)
					dc.Logger.Log("info", fmt.Sprintf("Started: %s\n  %s", name, ci))
				} else if mergedConfig.ExitBehavior == "abort" {
					if strings.Contains(name, "_default_") {
						dc.Logger.Log("debug", "Stop watching containers. Default container stopped.")
						return
					}
					dc.Logger.Log("info", fmt.Sprintf("Container: %s stopped by itself. Stopping the default container...", name))
					defaultCont := dc.getDefaultContainerID(names)
					if defaultCont == "" {
						dc.Logger.Log("debug", "Stop watching containers. Default container already removed")
						return
					}
					cmd := fmt.Sprintf("docker stop %s", defaultCont)
					stdout, stderr, exitStatus, _ := dc.ShellService.RunGetOutput(cmd, false)
					ci := cmdInfoToString(cmd, stdout, stderr, exitStatus)
					dc.Logger.Log("info", fmt.Sprintf("Stopped: %s.\n%s", defaultCont, ci))
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func safelyCloseChannel(ch chan bool) (justClosed bool) {
	defer func() {
		if recover() != nil {
			// The return result can be altered
			// in a defer function call.
			justClosed = false
		}
	}()

	// assume ch != nil here.
	close(ch)   // panic if ch is closed
	return true // <=> justClosed = true; return
}

func (dc DockerComposeDriver) getNonDefaultContainersInfos(containersNames []string) (map[string](map[string]string)) {
	containerInfos := make(map[string](map[string]string))
	for _, containerName := range containersNames {
		if strings.Contains(containerName, "_default_") {
			continue
		} else {
			infoHash := make(map[string]string)
			containerInfo, err := getContainerInfo(dc.ShellService, containerName)
			if err != nil {
				panic(err)
			}
			infoHash["status"] = containerInfo.Status
			infoHash["exitcode"] = containerInfo.ExitCode
			containerInfos[containerName] = infoHash
		}
	}
	return containerInfos
}

func (dc DockerComposeDriver) getNonDefaultContainersLogs(containersNames []string, containerInfos map[string](map[string]string)) map[string](map[string]string){
	for _, containerName := range containersNames {
		if strings.Contains(containerName, "_default_") {
			continue
		} else {
			cmd := fmt.Sprintf("docker logs %s", containerName)
			stdout, stderr, exitStatus, _ := dc.ShellService.RunGetOutput(cmd, true)
			if exitStatus != 0 {
				dc.Logger.Log("debug", fmt.Sprintf("Problem with getting containerInfos from: %s, problem: %s",
					containerName, stderr))
			}
			containerInfos[containerName]["logs"] = "stderr:\n" + stderr + "stdout:\n" + stdout
		}
	}
	return containerInfos
}

func checkIfAnyContainerFailed(nonDefaultContainerInfos map[string](map[string]string), defaultContainerExitCode int) bool {
	anyContainerFailed := false
	for _, v := range nonDefaultContainerInfos {
		if v["exitcode"] != "0" {
			anyContainerFailed = true
		}
	}
	anyContainerFailed = (defaultContainerExitCode != 0) || anyContainerFailed
	return anyContainerFailed
}


func (dc DockerComposeDriver) HandleRun(mergedConfig Config, runID string, envService EnvServiceInterface) int {
	warnGeneral(dc.FileService, mergedConfig, envService, dc.Logger)
	envFile, envFileMultiLine := getEnvFilePaths(runID, mergedConfig.Test)
	saveEnvToFile(dc.FileService, envFile, envFileMultiLine, mergedConfig.BlacklistVariables, envService.GetVariables())
	dojoDCGeneratedFile, err := dc.handleDCFiles(mergedConfig)
	if err != nil {
		return 1
	}
	expContainers := dc.getExpectedContainers(mergedConfig, runID)
	additionalContents := dc.generateDCFileContentsWithEnv(expContainers, mergedConfig, envFile, envFileMultiLine)
	dc.FileService.AppendContents(dojoDCGeneratedFile, additionalContents, "debug")

	cmd := dc.ConstructDockerComposeCommandRun(mergedConfig, runID)
	if isChannelClosed(dc.Stopping) {
		dc.Logger.Log("info", "Aborting containers start")
		return 0
	}

	dc.Logger.Log("info", green(fmt.Sprintf("docker-compose run command will be:\n %v", cmd)))
	go dc.watchContainers(mergedConfig, runID, len(expContainers))
	exitStatus, _ := dc.ShellService.RunInteractive(cmd, true)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from run command: %v", exitStatus))
	// Here:
	// * either "docker-compose run" finished by itself, we expect the default container to be stopped/removed. Let's stop the
	//   other containers.
	// * or it was stopped by a handle signal function, we expect all containers to be stopped (default container
	// may be removed)

	dc.Logger.Log("debug", fmt.Sprintf("Collecting information from other containers"))
	containersNames := dc.getDCContainersNames(mergedConfig, runID)
	containersInfos := dc.getNonDefaultContainersInfos(containersNames)
	anyContainerFailed := checkIfAnyContainerFailed(containersInfos, exitStatus)
	if mergedConfig.PrintLogs == "always" || (mergedConfig.PrintLogs == "failure" && anyContainerFailed) {
		dc.getNonDefaultContainersLogs(containersNames, containersInfos)
		for k, v := range containersInfos {
			containerInfo := v
			status := containerInfo["status"]
			if status == "running" {
				dc.Logger.Log("debug", fmt.Sprintf("Here are logs of container: %s, which status is: %s\n%s",
					k, status, containerInfo["logs"]))
			} else if status == "exited" {
				dc.Logger.Log("debug", fmt.Sprintf("Here are logs of container: %s, which exited with exitcode: %s\n%s",
					k, containerInfo["exitcode"], containerInfo["logs"]))
			} else {
				dc.Logger.Log("debug", fmt.Sprintf("Here are logs of container: %s, which status is: %s, exitcode: %s\n%s",
					k, status, containerInfo["exitcode"], containerInfo["logs"]))
			}
		}
	}

	dc.stop(mergedConfig, runID, "")
	return exitStatus

	// do not clean now, containers may be being stopped in other goroutines
}

func (dc DockerComposeDriver) CleanAfterRun(mergedConfig Config, runID string) int {
	if mergedConfig.RemoveContainers == "true" {
		dc.Logger.Log("debug", "Cleaning, because RemoveContainers is set to true")
		envFile, envFileMultiLine := getEnvFilePaths(runID, mergedConfig.Test)
		defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFile)
		defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFileMultiLine)
		dojoDCGeneratedFile := dc.getDCGeneratedFilePath(mergedConfig.DockerComposeFile)
		defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, dojoDCGeneratedFile)

		// Let's use docker-compose down instead of docker-compose rm command, because down also removes networks.
		cmd := dc.ConstructDockerComposeCommandDown(mergedConfig, runID)
		dc.Logger.Log("info", fmt.Sprintf("Removing containers with command: \n%v", cmd))
		exitStatus, _ := dc.ShellService.RunInteractive(cmd, true)
		return exitStatus
	} else {
		dc.Logger.Log("debug", "Not cleaning, because RemoveContainers is not set to true")
		return 0
	}
}

func (dc DockerComposeDriver) stop(mergedConfig Config, runID string, defaultContainerID string) int {
	if isChannelClosed(dc.Stopping) {
		// either HandleRun() or HandleSignal() already scheduled stopping
		dc.Logger.Log("debug", "Containers are already being stopped, nothing more to do")
		return 0
	} else {
		dc.Logger.Log("debug", "Stopping containers")
		safelyCloseChannel(dc.Stopping)
	}
	if defaultContainerID != "" {
		cmd := fmt.Sprintf("docker stop %s", defaultContainerID)
		dc.Logger.Log("info", fmt.Sprintf("Stopping default container with command: \n%v", cmd))
		exitStatus, _ := dc.ShellService.RunInteractive(cmd, false)
		if exitStatus != 0 {
			dc.Logger.Log("error", fmt.Sprintf("Unexpected exit status: %v from stop command: %s", exitStatus, cmd))
		}
	} // else container was stopped & removed already

	// "docker-compose stop" command stops only the non-default containers, thus we stopped the default
	// container separately above
	cmd := dc.ConstructDockerComposeCommandStop(mergedConfig, runID)
	dc.Logger.Log("info", fmt.Sprintf("Stopping containers with command: \n%v", cmd))
	exitStatus, _ := dc.ShellService.RunInteractive(cmd, false)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from stop command: %v", exitStatus))
	return exitStatus
}

func (dc DockerComposeDriver) handleDCFiles(mergedConfig Config) (string, error) {
	fileContents := dc.FileService.ReadDockerComposeFile(mergedConfig.DockerComposeFile)
	version, err := dc.verifyDCFile(fileContents, mergedConfig.DockerComposeFile)
	if err != nil {
		dc.Logger.Log("error", fmt.Sprintf("Docker-compose file %s is not correct: %s", mergedConfig.DockerComposeFile, err.Error()))
		return "", err
	}
	dojoDCFileContents := dc.generateInitialDCFile(mergedConfig, version)
	dojoDCFileName := mergedConfig.DockerComposeFile + ".dojo"
	dc.FileService.WriteToFile(dojoDCFileName, dojoDCFileContents, "debug")
	return dojoDCFileName, nil
}

func (dc DockerComposeDriver) ConstructDockerComposeCommandPull(config Config, dojoGeneratedDCFile string) string {
	// projectName does not matter for docker-compose pull command
	cmd := dc.ConstructDockerComposeCommandPart1(config, "dojo")
	cmd += " pull"
	return cmd
}

func (dc DockerComposeDriver) HandlePull(mergedConfig Config) int {
	dojoDCGeneratedFile, err := dc.handleDCFiles(mergedConfig)
	if err != nil {
		return 1
	}
	defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, dojoDCGeneratedFile)

	cmd := dc.ConstructDockerComposeCommandPull(mergedConfig, dojoDCGeneratedFile)
	dc.Logger.Log("info", green(fmt.Sprintf("docker-compose pull command will be:\n %v", cmd)))
	exitStatus, _ := dc.ShellService.RunInteractive(cmd, false)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from pull command: %v", exitStatus))
	return exitStatus
}

// Stop the containers if stoppinng was not already started (channel closed).
func (dc DockerComposeDriver) HandleSignal(mergedConfig Config, runID string) int {
	dc.Logger.Log("info", "Stopping on signal")
	if isChannelClosed(dc.Stopping) {
		dc.Logger.Log("info", "Containers are already being stopped, will not react on this signal")
		return 0
	}

	names := dc.getDCContainersNames(mergedConfig, runID)
	if len(names) == 0 {
		dc.Logger.Log("info", fmt.Sprintf("Stopping not needed, the containers were not created: %s", runID))
		return 0
	}

	defaultContainerID := dc.getDefaultContainerID(names)
	es := dc.stop(mergedConfig, runID, defaultContainerID)
	dc.Logger.Log("info", "Stopping on signal finished")
	return es
}

func (dc DockerComposeDriver) ConstructDockerComposeCommandKill(mergedConfig Config, runID string) string {
	cmd := dc.ConstructDockerComposeCommandPart1(mergedConfig, runID)
	cmd += " kill"
	return cmd
}

func (dc DockerComposeDriver) kill(mergedConfig Config, runID string, defaultContainerID string) int {
	if defaultContainerID != "" {
		cmd := fmt.Sprintf("docker kill %s", defaultContainerID)
		exitStatus, _ := dc.ShellService.RunInteractive(cmd, true)
		if exitStatus != 0 {
			dc.Logger.Log("error", fmt.Sprintf("Exit status from stop command: %v", exitStatus))
		}
	} // else container was killed & removed already

	// "docker-compose kill" command stops only the non-default containers, thus we killed the default
	// container separately above
	cmd := dc.ConstructDockerComposeCommandKill(mergedConfig, runID)
	dc.Logger.Log("debug", fmt.Sprintf("Stopping containers with command: \n%v", cmd))
	exitStatus, _ := dc.ShellService.RunInteractive(cmd, true)
	dc.Logger.Log("debug", fmt.Sprintf("Exit status from stop command: %v", exitStatus))
	return exitStatus
}

// Kill the containers.
func (dc DockerComposeDriver) HandleMultipleSignal(mergedConfig Config, runID string) int {
	dc.Logger.Log("info", "Stopping on multiple signals")

	names := dc.getDCContainersNames(mergedConfig, runID)
	if len(names) == 0 {
		dc.Logger.Log("info", fmt.Sprintf("Stopping not needed, the containers were not created: %s", runID))
		return 0
	}

	defaultContainerID := dc.getDefaultContainerID(names)
	es := dc.kill(mergedConfig, runID, defaultContainerID)
	dc.Logger.Log("info", "Stopping on multiple signals finished")
	return es
}


func (dc DockerComposeDriver) getDefaultContainerID(containersNames []string) string {
	for _, containerName := range containersNames {
		if strings.Contains(containerName, "_default_") {
			contanerInfo, err := getContainerInfo(dc.ShellService, containerName)
			if !contanerInfo.Exists {
				return ""
			}
			if err != nil {
				panic(err)
			}
			dc.Logger.Log("debug", fmt.Sprintf("Found default container ID: %s", contanerInfo.ID))
			return contanerInfo.ID
		}
	}
	panic(fmt.Errorf("default container not found. Were the containers created?"))
}

// example output of docker-compose ps:
//Name                        Command               State   Ports
//------------------------------------------------------------------------
//edudocker_abc_1           /bin/sh -c while true; do  ...   Up
//edudocker_def_1           /bin/sh -c while true; do  ...   Up
//edudocker_default_run_1   sh -c echo 'will sleep' && ...   Up
// Returns: container names, error
func (dc DockerComposeDriver) getDCContainersNames(mergedConfig Config, projectName string) ([]string) {
	if projectName == "" {
		panic("projectName was empty")
	}
	cmd := dc.ConstructDockerComposeCommandPs(mergedConfig, projectName)
	stdout, stderr, exitStatus, _ := dc.ShellService.RunGetOutput(cmd, true)
	if exitStatus != 0 {
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
		panic(fmt.Errorf("Unexpected exit status:\n%s", cmdInfo))
	}
	stdout = strings.TrimSuffix(stdout, "\n")
	if stdout == "" {
		// this happens only in unit tests, when I forget to mock docker-compose ps command
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
		panic(fmt.Errorf("Unexpected empty stdout: %s", cmdInfo))
	}
	lines := strings.Split(stdout,"\n")
	if strings.Contains(lines[len(lines)-1], "-----") {
		// the last line of output is the one with -----
		dc.Logger.Log("debug", "Containers were not yet created")
		return []string{}
	}
	containerLines := lines[2:]
	containersNames := make([]string, 0)
	for _, line := range containerLines {
		split := strings.Split(line, " ")
		containerName := split[0]
		containersNames = append(containersNames, containerName)
	}
	return containersNames
}
