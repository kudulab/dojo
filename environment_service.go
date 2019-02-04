package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

type EnvServiceInterface interface {
	Variables() []string
	IsCurrentUserRoot() bool
}

type EnvService struct {}

func (f EnvService) Variables() []string {
	return os.Environ()
}
func (f EnvService) IsCurrentUserRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	if currentUser.Username == "root" {
		return true
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
func saveEnvToFile(fileService FileServiceInterface, envFilePath string, blacklistedVars string, currentVariables []string)  {
	if fileService == nil {
		panic("fileService was nil")
	}
	fileService.RemoveFile(envFilePath, true)
	fileContents := generateVariablesString(blacklistedVars, currentVariables)
	fileService.WriteToFile(envFilePath, fileContents, "debug")
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
			generatedString += "DISPLAY=unix:0.0\n"
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

func getEnvFilePath(runID string, test string) string {
	if test == "true" {
		return fmt.Sprintf("/tmp/test-dojo-environment-%s", runID)
	} else {
		return fmt.Sprintf("/tmp/dojo-environment-%s", runID)
	}
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