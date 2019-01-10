package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func parseDCFileVersion(contents string) (float64, error) {
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

func verifyDCFile(fileContents string, filePath string) (float64, error) {
	version, err := parseDCFileVersion(fileContents)
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

func handleDCFiles(mergedConfig Config, envFile string) (string) {
	contents, err := ioutil.ReadFile(mergedConfig.DockerComposeFile)
	if err != nil {
		panic(err)
	}
	fileContents := string(contents)
	version, err := verifyDCFile(fileContents, mergedConfig.DockerComposeFile)
	if err != nil {
		panic(err)
	}
	dojoDCFileContents := generateDCFileContents(mergedConfig, version, envFile)
	dojoDCFileName := mergedConfig.DockerComposeFile + ".dojo"
	Log("debug", fmt.Sprintf("docker-compose dojo file contents will be:\n%s", dojoDCFileContents))
	if mergedConfig.Dryrun != "true" {
		err = ioutil.WriteFile(dojoDCFileName, []byte(dojoDCFileContents), 0664)
		if err != nil {
			panic(err)
		}
		Log("debug", fmt.Sprintf("Created docker-compose dojo file: %v", dojoDCFileName))
	} else {
		Log("debug", fmt.Sprintf("Not created docker-compose dojo file: %v, because  dryrun is set", dojoDCFileName))
	}
	return dojoDCFileName
}

func generateDCFileContents(config Config, version float64, envFile string) string {
	if config.Dryrun != "true" {
		if getFileUid(config.WorkDirOuter) == 0 {
			Log("warn", fmt.Sprintf("WorkDirOuter: %s is owned by root, which is not recommended", config.WorkDirOuter))
		}
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

func constructDockerComposeCommandPart1(config Config, projectName string, dojoGeneratedDCFile string) string {
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
func constructDockerComposeCommandRun(config Config, projectName string, dojoGeneratedDCFile string, shellIsInteractive bool) string {
	cmd := constructDockerComposeCommandPart1(config, projectName, dojoGeneratedDCFile)
	cmd += " run --rm"
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
func constructDockerComposeCommandStop(config Config, projectName string, dojoGeneratedDCFile string) string {
	cmd := constructDockerComposeCommandPart1(config, projectName, dojoGeneratedDCFile)
	cmd += " stop"
	return cmd
}
func constructDockerComposeCommandRm(config Config, projectName string, dojoGeneratedDCFile string) string {
	cmd := constructDockerComposeCommandPart1(config, projectName, dojoGeneratedDCFile)
	cmd += " rm -f"
	return cmd
}
func getExpDockerNetwork(runID string) string {
	// remove dashes and underscores and add "_default" (we always demand docker container named: "default")
	runIDPrim := strings.Replace(runID, "/","", -1)
	runIDPrim = strings.Replace(runIDPrim, "_","", -1)
	runIDPrim = strings.Replace(runIDPrim, "-","", -1)
	network := fmt.Sprintf("%s_default", runIDPrim)
	return network
}