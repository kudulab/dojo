package main

import (
	"encoding/json"
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
	Logger       *Logger
	// This channel is closed when the stop action is started
	Stopping             chan bool
	DockerComposeVersion string
}

func NewDockerComposeDriver(shellService ShellServiceInterface, fs FileServiceInterface, logger *Logger, version string) DockerComposeDriver {
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
		ShellService:         shellService,
		FileService:          fs,
		Logger:               logger,
		Stopping:             stopping,
		DockerComposeVersion: version,
	}
}

func GetDockerComposeVersion(shellService ShellServiceInterface) string {
	cmd := "docker-compose version --short"
	stdout, stderr, es, _ := shellService.RunGetOutput(cmd, true)
	if es != 0 || stderr != "" || stdout == "" {
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, es)
		panic(fmt.Errorf("Unexpected error: %s", cmdInfo))
	}
	stdout = strings.TrimSuffix(stdout, "\n")
	return stdout
}

func isDCVersionLaterThan2(dcVersion string) bool {
	if strings.HasPrefix(dcVersion, "2") || strings.HasPrefix(dcVersion, "v2") {
		return true
	}
	return false
}

func (dc DockerComposeDriver) parseDCFileVersion(contents string) (float64, error) {
	firstLine := strings.Split(contents, "\n")[0]
	if !strings.HasPrefix(firstLine, "version") {
		return -1, nil
	}
	versionQuoted := strings.Split(firstLine, ":")[1]
	versionQuoted = strings.Trim(versionQuoted, " ")
	versionQuoted = strings.TrimSuffix(versionQuoted, "\n")
	versionQuoted = strings.Trim(versionQuoted, "\"")
	versionQuoted = strings.Trim(versionQuoted, "'")
	version, err := strconv.ParseFloat(versionQuoted, 64)
	if err != nil {
		return 0, err
	}
	return version, nil
}
func (dc DockerComposeDriver) verifyDCFile(fileContents string, filePath string) (float64, error) {
	version, err := dc.parseDCFileVersion(fileContents)
	if err != nil {
		return 0, err
	}
	if version != -1 && (version < 2 || version >= 3) {
		return 0, fmt.Errorf("docker-compose file: %s should contain version number >=2 and <3, current version: %v", filePath, version)
	}
	requiredStr := "default:"
	if !strings.Contains(fileContents, requiredStr) {
		return 0, fmt.Errorf("docker-compose file: %s does not contain: %s. Please add a default service", filePath, requiredStr)
	}
	return version, nil
}

func (dc DockerComposeDriver) getDCGeneratedFilePath(dcfilePath string) string {
	return dcfilePath + ".dojo"
}

func (dc DockerComposeDriver) generateDCFileContentsWithEnv(expContainers []string, config Config, envFile string,
	envFileMultiLine string, envFileBashFunctions string) string {
	contents := fmt.Sprintf(
		`    volumes:
      - %s:%s:ro
      - %s:%s
      - %s:/etc/dojo.d/variables/00-multiline-vars.sh
      - %s:/etc/dojo.d/variables/01-bash-functions.sh
`, config.IdentityDirOuter, "/dojo/identity", config.WorkDirOuter, config.WorkDirInner, envFileMultiLine, envFileBashFunctions)
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
      - %s:/etc/dojo.d/variables/01-bash-functions.sh
`, name, envFile, envFileMultiLine, envFileBashFunctions)
		}
	}
	return contents
}

func (dc DockerComposeDriver) generateInitialDCFile(config Config, version float64) string {
	if version == -1 {
		// version was not set
		return fmt.Sprintf(
			`services:
  default:
    image: "%s"`, config.DockerImage)
	}
	return fmt.Sprintf(
		`version: '%v'
services:
  default:
    image: "%s"`, version, config.DockerImage)
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
	} else if config.Interactive == "true" {
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
	if isDCVersionLaterThan2(dc.DockerComposeVersion) {
		// We need to all the option `--all` to include the output about the default container.
		// We can use the json format now. It was unavailable in previous versions.
		cmd += " --format json --all"
	}
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
	case <-ch:
		return true
	default:
	}
	return false
}

func (dc DockerComposeDriver) checkAllContainersRunning(containerIDs []string) bool {
	for _, id := range containerIDs {
		running := dc.checkContainerIsRunning(id)
		if !running {
			return false
		}
	}
	return true
}

func (dc DockerComposeDriver) checkContainerIsRunning(containerID string) bool {
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
func (dc DockerComposeDriver) waitForContainersToBeRunning(mergedConfig Config, runID string, expContainersCount int) []string {
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
					if strings.Contains(name, "_default_") || strings.Contains(name, "-default-") {
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

func (dc DockerComposeDriver) getNonDefaultContainersInfos(containersNames []string) []*ContainerInfo {
	containerInfos := make([]*ContainerInfo, 0)
	for _, containerName := range containersNames {
		if strings.Contains(containerName, "_default_") || strings.Contains(containerName, "-default-") {
			continue
		} else {
			containerInfo, err := getContainerInfo(dc.ShellService, containerName)
			if err != nil {
				panic(err)
			}
			containerInfos = append(containerInfos, containerInfo)
		}
	}
	return containerInfos
}

func (dc DockerComposeDriver) getNonDefaultContainersLogs(containerInfos []*ContainerInfo) {
	for _, containerInfo := range containerInfos {
		containerName := containerInfo.Name
		if strings.Contains(containerName, "_default_") || strings.Contains(containerName, "-default-") {
			continue
		} else {
			cmd := fmt.Sprintf("docker logs %s", containerName)
			stdout, stderr, exitStatus, _ := dc.ShellService.RunGetOutput(cmd, true)
			if exitStatus != 0 {
				dc.Logger.Log("debug", fmt.Sprintf("Problem with getting containerInfos from: %s, problem: %s",
					containerName, stderr))
			}
			containerInfo.Logs = "stderr:\n" + stderr + "stdout:\n" + stdout
		}
	}
}

func checkIfAnyContainerFailed(nonDefaultContainerInfos []*ContainerInfo, defaultContainerExitCode int) bool {
	anyContainerFailed := false
	for _, v := range nonDefaultContainerInfos {
		if v.ExitCode != "0" {
			anyContainerFailed = true
		}
	}
	anyContainerFailed = (defaultContainerExitCode != 0) || anyContainerFailed
	return anyContainerFailed
}

func (d DockerComposeDriver) PrintVersion() {
	version_cmd := "docker-compose --version"
	stdout, stderr, exitStatus, _ := d.ShellService.RunGetOutput(version_cmd, true)
	if exitStatus != 0 {
		cmdInfo := cmdInfoToString(version_cmd, stdout, stderr, exitStatus)
		d.Logger.Log("debug", cmdInfo)
	} else {
		d.Logger.Log("info", stdout)
	}
}

func (dc DockerComposeDriver) HandleRun(mergedConfig Config, runID string, envService EnvServiceInterface) int {
	warnGeneral(dc.FileService, mergedConfig, envService, dc.Logger)
	envFile, envFileMultiLine, envFileBashFunctions := getEnvFilePaths(runID, mergedConfig.Test)
	saveEnvToFile(dc.FileService, envFile, envFileMultiLine, envFileBashFunctions,
		mergedConfig.BlacklistVariables, envService.GetVariables())
	dojoDCGeneratedFile, err := dc.handleDCFiles(mergedConfig)
	if err != nil {
		return 1
	}
	expContainers := dc.getExpectedContainers(mergedConfig, runID)
	additionalContents := dc.generateDCFileContentsWithEnv(expContainers, mergedConfig, envFile, envFileMultiLine, envFileBashFunctions)
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

	dc.Logger.Log("debug", fmt.Sprintf("Collecting information from non default containers"))
	containersNames := dc.getDCContainersNames(mergedConfig, runID)
	containersInfos := dc.getNonDefaultContainersInfos(containersNames)
	anyContainerFailed := checkIfAnyContainerFailed(containersInfos, exitStatus)
	if mergedConfig.PrintLogs == "always" || (mergedConfig.PrintLogs == "failure" && anyContainerFailed) {
		dc.Logger.Log("debug", fmt.Sprintf("Getting non default containers logs"))
		dc.getNonDefaultContainersLogs(containersInfos)
		dc.Logger.Log("debug", fmt.Sprintf("Got logs from %s containers", fmt.Sprint(len(containersInfos))))
		for _, v := range containersInfos {
			containerInfo := v
			status := containerInfo.Status
			statusMsg := ""
			if status == "running" {
				statusMsg = fmt.Sprintf("which status is: %s", status)
			} else if status == "exited" {
				statusMsg = fmt.Sprintf("which exited with exitcode: %s", containerInfo.ExitCode)
			} else {
				statusMsg = fmt.Sprintf("which status is: %s, exitcode: %s", status, containerInfo.ExitCode)
			}

			if mergedConfig.PrintLogsTarget == "file" {
				logsFilePath := "dojo-logs-" + containerInfo.Name + "-" + runID + ".txt"
				dc.FileService.WriteToFile(logsFilePath, containerInfo.Logs, "debug")
				dc.Logger.Log("info", fmt.Sprintf("The logs of container: %s, %s, were saved to file: %s",
					containerInfo.Name, statusMsg, logsFilePath))
			} else {
				dc.Logger.Log("info", fmt.Sprintf("Here are logs of container: %s, %s:\n%s",
					containerInfo.Name, statusMsg, containerInfo.Logs))
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
		envFile, envFileMultiLine, envFilePathBashFunctions := getEnvFilePaths(runID, mergedConfig.Test)
		defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFile)
		defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFileMultiLine)
		defer dc.FileService.RemoveGeneratedFile(mergedConfig.RemoveContainers, envFilePathBashFunctions)
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
		if strings.Contains(containerName, "_default_") || strings.Contains(containerName, "-default-") {
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

// Example output of `docker-compose ps` command, when using docker-compose <2:
// Name                        Command               State   Ports
// ------------------------------------------------------------------------
// edudocker_abc_1           /bin/sh -c while true; do  ...   Up
// edudocker_def_1           /bin/sh -c while true; do  ...   Up
// edudocker_default_run_1   sh -c echo 'will sleep' && ...   Up
//
// All the containers, including the default one, are included in the output.
//
// Example output of `docker-compose ps` command, when using docker-compose >2:
// NAME                  IMAGE         COMMAND                  SERVICE   CREATED         STATUS         PORTS
// testdojorunid-abc-1   alpine:3.19   "/bin/sh -c 'while t…"   abc       6 seconds ago   Up 5 seconds
//
// The output for docker-compose >2 does not show the default container. In order to have it included, we need to run `ps --all`
//
// Returns: container names
func (dc DockerComposeDriver) getDCContainersNames(mergedConfig Config, projectName string) []string {
	if projectName == "" {
		panic("projectName was empty")
	}
	cmd := dc.ConstructDockerComposeCommandPs(mergedConfig, projectName)

	stdout, stderr, exitStatus, _ := dc.ShellService.RunGetOutput(cmd, true)
	if exitStatus != 0 {
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
		if strings.Contains(stderr, "No such container") {
			// Workaround for issue #27: do not panic but print error level log message.
			// This happens rarely and seems similar to
			// https://github.com/docker/compose/issues/9534 and
			// https://github.com/docker/compose/issues/10373
			// and this makes the tests such as test_docker_compose_run_preserves_bash_functions flaky
			dc.Logger.Log("error", fmt.Sprintf("Unexpected exit status:\n%s", cmdInfo))
			return []string{}
		} else {
			panic(fmt.Errorf("Unexpected exit status:\n%s", cmdInfo))
		}
	}

	if isDCVersionLaterThan2(dc.DockerComposeVersion) {
		jsonOutputArr, err := ParseDCPSOutPut_DCVersion2(stdout)
		if err != "" {
			// It seems that nothing in this function needs to panic.
			// Even if there are errors, the main functionality could still work, but additional features
			// (such as printing logs to file) would be affected.
			dc.Logger.Log("error", err)
			return []string{}
		}
		containersNames := make([]string, 0)
		for _, container := range jsonOutputArr {
			containersNames = append(containersNames, container.Name)
		}
		return containersNames
	}

	// getting container names while using docker-compose <2

	stdout = strings.TrimSuffix(stdout, "\n")
	if stdout == "" {
		// this happens only in unit tests, when I forget to mock docker-compose ps command
		cmdInfo := cmdInfoToString(cmd, stdout, stderr, exitStatus)
		panic(fmt.Errorf("Unexpected empty stdout: %s", cmdInfo))
	}
	lines := strings.Split(stdout, "\n")
	if len(lines) < 2 || strings.Contains(lines[len(lines)-1], "-----") {
		// if running on Linux or using Docker Desktop,
		// the last line of output is the one with -----;
		// if using Colima, there is just 1 line of output
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

type DC2PSOutput struct {
	Command    string   `json:"Command"`
	CreatedAt  string   `json:"CreatedAt"`
	ExitCode   int      `json:"ExitCode"`
	Health     string   `json:"Health"`
	ID         string   `json:"ID"`
	Image      string   `json:"Image"`
	Labels     string   `json:"Labels"`
	Name       string   `json:"Name"`
	Names      string   `json:"Names"`
	Networks   string   `json:"Networks"`
	Ports      string   `json:"Ports"`
	Project    string   `json:"Project"`
	Publishers []string `json:"Publishers"`
	RunningFor string   `json:"RunningFor"`
	Service    string   `json:"Service"`
	Size       string   `json:"Size"`
	State      string   `json:"State"`
	Status     string   `json:"Status"`
}

// Parses the output of the `docker-compose ps --format json --all` command, into json.
//
// Example of running the ps command, when using docker-compose: 2.24.5:
// $ docker-compose -f ./test/test-files/itest-dc.yaml -f ./test/test-files/itest-dc.yaml.dojo -p testdojorunid ps --format json --all
// {"Command":"\"/bin/sh -c 'while t…\"","CreatedAt":"2024-02-03 21:03:46 +0000 UTC","ExitCode":0,"Health":"","ID":"2d5c5b0343d0","Image":"alpine:3.19","Labels":"com.docker.compose.depends_on=,com.docker.compose.image=sha256:05455a08881ea9cf0e752bc48e61bbd71a34c029bb13df01e40e3e70e0d007bd,com.docker.compose.version=2.24.5,com.docker.compose.service=abc,com.docker.compose.config-hash=270e27422cb1e6a4c1713ae22a3ffca0e8aa50ec0f06fe493fa4f83a17bd29e9,com.docker.compose.container-number=1,com.docker.compose.oneoff=False,com.docker.compose.project=testdojorunid,com.docker.compose.project.config_files=/dojo/work/test/test-files/itest-dc.yaml,/dojo/work/test/test-files/itest-dc.yaml.dojo,com.docker.compose.project.working_dir=/dojo/work/test/test-files","LocalVolumes":"0","Mounts":"/tmp/test-dojo…,/tmp/test-dojo…","Name":"testdojorunid-abc-1","Names":"testdojorunid-abc-1","Networks":"testdojorunid_default","Ports":"","Project":"testdojorunid","Publishers":null,"RunningFor":"3 seconds ago","Service":"abc","Size":"0B","State":"running","Status":"Up 2 seconds"}
// {"Command":"\"/bin/sh -c 'while t…\"","CreatedAt":"2024-02-03 21:03:46 +0000 UTC","ExitCode":0,"Health":"","ID":"b2ed210567c3","Image":"alpine:3.19","Labels":"com.docker.compose.project=testdojorunid,com.docker.compose.project.config_files=/dojo/work/test/test-files/itest-dc.yaml,/dojo/work/test/test-files/itest-dc.yaml.dojo,com.docker.compose.project.working_dir=/dojo/work/test/test-files,com.docker.compose.depends_on=,com.docker.compose.container-number=1,com.docker.compose.image=sha256:05455a08881ea9cf0e752bc48e61bbd71a34c029bb13df01e40e3e70e0d007bd,com.docker.compose.oneoff=False,com.docker.compose.service=def,com.docker.compose.version=2.24.5,com.docker.compose.config-hash=270e27422cb1e6a4c1713ae22a3ffca0e8aa50ec0f06fe493fa4f83a17bd29e9","LocalVolumes":"0","Mounts":"/tmp/test-dojo…,/tmp/test-dojo…","Name":"testdojorunid-def-1","Names":"testdojorunid-def-1","Networks":"testdojorunid_default","Ports":"","Project":"testdojorunid","Publishers":null,"RunningFor":"3 seconds ago","Service":"def","Size":"0B","State":"running","Status":"Up 2 seconds"}
// {"Command":"\"sh -c 'sleep 10'\"","CreatedAt":"2024-02-03 21:03:47 +0000 UTC","ExitCode":0,"Health":"","ID":"af4817fede41","Image":"alpine:3.15","Labels":"com.docker.compose.version=2.24.5,com.docker.compose.container-number=1,com.docker.compose.depends_on=abc:service_started:true,def:service_started:true,com.docker.compose.oneoff=True,com.docker.compose.project=testdojorunid,com.docker.compose.slug=742bcbb0e4bc05b21928a8d17be4ea9bb12a6775fd40692dd59c74a460279eb8,com.docker.compose.config-hash=462afacb4521d13580c2096c7b00b98970f07fe841e408c4c5a95a4a46839eaa,com.docker.compose.image=sha256:32b91e3161c8fc2e3baf2732a594305ca5093c82ff4e0c9f6ebbd2a879468e1d,com.docker.compose.project.config_files=/dojo/work/test/test-files/itest-dc.yaml,/dojo/work/test/test-files/itest-dc.yaml.dojo,com.docker.compose.project.working_dir=/dojo/work/test/test-files,com.docker.compose.service=default","LocalVolumes":"0","Mounts":"/home/dojo,/dojo/work,/tmp/test-dojo…,/tmp/test-dojo…,/tmp/.X11-unix,/tmp/dojo-ites…","Name":"testdojorunid-default-run-742bcbb0e4bc","Names":"testdojorunid-default-run-742bcbb0e4bc","Networks":"testdojorunid_default","Ports":"","Project":"testdojorunid","Publishers":null,"RunningFor":"2 seconds ago","Service":"default","Size":"0B","State":"running","Status":"Up 1 second"}
//
// when using docker-compose: 2.31m the field with Publishers is set to []:
// {"Command":"\"/bin/sh -c 'while t…\"","CreatedAt":"2024-12-20 15:58:00 +1300 NZDT","ExitCode":143,"Health":"","ID":"f543828473a7","Image":"alpine:3.19","Labels":"com.docker.compose.image=sha256:7a85bf5dc56c949be827f84f9185161265c58f589bb8b2a6b6bb6d3076c1be21,desktop.docker.io/binds/0/Source=/tmp/dojo-environment-multiline-dojo-test-files-2024-12-2015-58-00-6660523964295751458,desktop.docker.io/binds/0/SourceKind=hostFile,desktop.docker.io/binds/0/Target=/etc/dojo.d/variables/00-multiline-vars.sh,desktop.docker.io/binds/1/Target=/etc/dojo.d/variables/01-bash-functions.sh,com.docker.compose.oneoff=False,com.docker.compose.project.config_files=/Users/ava.czechowska/code/dojo/test/test-files/itest-dc.yaml,/Users/ava.czechowska/code/dojo/test/test-files/itest-dc.yaml.dojo,com.docker.compose.service=abc,desktop.docker.io/binds/1/SourceKind=hostFile,com.docker.compose.container-number=1,com.docker.compose.project.working_dir=/Users/ava.czechowska/code/dojo/test/test-files,com.docker.compose.config-hash=4439e404114c9d0f183d756084f3e0e80c56b731e1f1e14e6db1844134294969,com.docker.compose.depends_on=,com.docker.compose.project=dojo-test-files-2024-12-2015-58-00-6660523964295751458,com.docker.compose.version=2.31.0,desktop.docker.io/binds/1/Source=/tmp/dojo-environment-bash-functions-dojo-test-files-2024-12-2015-58-00-6660523964295751458","LocalVolumes":"0","Mounts":"/host_mnt/priv…,/host_mnt/priv…","Name":"dojo-test-files-2024-12-2015-58-00-6660523964295751458-abc-1","Names":"dojo-test-files-2024-12-2015-58-00-6660523964295751458-abc-1","Networks":"dojo-test-files-2024-12-2015-58-00-6660523964295751458_default","Ports":"","Project":"dojo-test-files-2024-12-2015-58-00-6660523964295751458","Publishers":[],"RunningFor":"36 seconds ago","Service":"abc","Size":"0B","State":"exited","Status":"Exited (143) 10 seconds ago"}
func ParseDCPSOutPut_DCVersion2(output string) ([]DC2PSOutput, string) {
	var output_as_jsons []DC2PSOutput

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var one_output_as_json DC2PSOutput
		err := json.Unmarshal([]byte(line), &one_output_as_json)
		if err != nil {
			return []DC2PSOutput{},
				fmt.Errorf("Error when decoding the JSON response from docker-compose ps command: %s; line %s", err, line).Error()
		}
		if one_output_as_json.State == "" {
			// This means that something went wrong, e.g. docker-compose ps output is now different
			// than expected.
			return []DC2PSOutput{},
				fmt.Errorf("Error when decoding the JSON response from docker-compose ps command: container State was an empty string. More details: %#v", one_output_as_json).Error()
		}
		output_as_jsons = append(output_as_jsons, one_output_as_json)
	}

	return output_as_jsons, ""
}
