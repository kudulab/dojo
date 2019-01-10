package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"
)

func getFileUid(filePath string) uint32 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}
	mode := fileInfo.Sys().(*syscall.Stat_t)
	uid := mode.Uid
	return uid
}

func constructDockerCommand(config Config, envFilePath string, containerName string, shellIsInteractive bool) string {
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
	if config.Dryrun != "true" {
		if getFileUid(config.WorkDirOuter) == 0 {
			Log("warn", fmt.Sprintf("WorkDirOuter: %s is owned by root, which is not recommended", config.WorkDirOuter))
		}
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

// Returns an identificator that can be reused later in many places,
// e.g. as some file name or as docker container name.
// e.g. dojo-myproject-2019-01-09_10-39-06-98498093
// It may not contain upper case letters or else "docker inspect" complains with the error:
// invalid reference format: repository name must be lowercase.
func getRunID() string {
	currentDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	currentDirectorySplit := strings.Split(currentDirectory, "/")
	currentDirectoryLastPart := currentDirectorySplit[len(currentDirectorySplit)-1]

	currentTime := time.Now().Format("2006-01-02_15-04-05")
	// run ID must contain a random number. Using time is insufficient, because e.g. 2 CI agents may be started
	// in the same second for the same project.
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(99999999)
	return fmt.Sprintf("dojo-%s-%v-%v", currentDirectoryLastPart, currentTime, randomNumber)
}

func isVariableBlacklisted(variableName string, blacklistedVariables []string) bool {
	for _,v := range blacklistedVariables {
		if strings.HasSuffix(v, "*") {
			vNoSuffix := strings.TrimSuffix(v, "*")
			if strings.HasPrefix(variableName, vNoSuffix) {
				return true
			}
		}
		if variableName == v {
			return true
		}
	}
	return false
}

// true if e.g. USER and DOJO_USER are set
func existsVariableWithDOJOPrefix(variableName string, allVariables []string) bool {
	for _,v := range allVariables {
		arr := strings.SplitN(v,"=", 2)
		key := arr[0]

		if strings.HasPrefix(key, "DOJO_") {
			vNoPrefix := strings.TrimPrefix(key, "DOJO_")
			if vNoPrefix == variableName {
				return true
			}
		}
	}
	return false
}

// Writes environment variables that will be preserved into a docker container to a file (envFilePath).
// Format is: ENV_VAR_NAME="env var value".
// Blacklisted variables are respected. If any env variable is blacklisted, it will be saved with "DOJO_" prefix.
// If env var with "DOJO_" prefix already exists, its value is taken, instead of
// the primary variable. E.g. PWD is blacklisted, so it will be saved as "DOJO_PWD=/some/path".
// If DOJO_PWD already exists, it is preserved as is and we do nothing to preserve PWD value.
// Variables can be also blacklisted with asterisk, e.g. BASH*. This means that
// any variable starting with BASH will be blacklisted (and prefixed).
// Variables with DOJO_ prefix cannot be blacklisted.
func saveEnvToFile(envFilePath string, blacklistedVars string, dryrun string) {
	if envFilePath == "" {
		panic("envFilePath was an empty string")
	}
	if dryrun != "true" {
		removeFile(envFilePath, true)
	}

	allVariables := os.Environ()
	fileContents := generateVariablesString(blacklistedVars, allVariables)
	if dryrun != "true" {
		err := ioutil.WriteFile(envFilePath, []byte(fileContents), 0644)
		if err != nil {
			panic(err)
		}
		Log("debug", fmt.Sprintf("Saved environment variables to file: %v", envFilePath))
	} else {
		Log("debug", fmt.Sprintf("Not saved environment variables to file: %v, because  dryrun is set", envFilePath))
	}
}

// allVariables is a []string, where each element is of format: VariableName=VariableValue
func generateVariablesString(blacklistedVarsNames string, allVariables []string) string {
	blacklistedVarsArr := strings.Split(blacklistedVarsNames, ",")
	generatedString := ""
	for _,v := range allVariables {
		arr := strings.SplitN(v,"=", 2)
		key := arr[0]
		value := arr[1]
		if key == "DISPLAY" {
			// this is highly opinionated
			generatedString += "DISPLAY=unix:0.0"
		} else if existsVariableWithDOJOPrefix(key, allVariables) {
			// ignore this key, we will deal with DOJO_${key}
			continue
		} else if strings.HasPrefix(key, "DOJO_") {
			generatedString += fmt.Sprintf("%s=%s\n", key, value)
		} else if isVariableBlacklisted(key, blacklistedVarsArr) {
			generatedString += fmt.Sprintf("DOJO_%s=%s\n", key, value)
		} else {
			generatedString += fmt.Sprintf("%s=%s\n", key, value)
		}
	}
	return generatedString
}