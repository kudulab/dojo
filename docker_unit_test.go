package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func getTestConfig() Config {
	config := getDefaultConfig("somefile")
	config.DockerImage = "img:1.2.3"
	// set these to some dummy dir, so that tests work also if not run in dojo docker image
	config.WorkDirOuter = "/tmp/bla"
	config.IdentityDirOuter = "/tmp/myidentity"
	config.Dryrun = "true"
	return config
}

func setTestEnv() {
	os.Unsetenv("DISPLAY")
}

func Test_ConstructDockerCommand_Interactive(t *testing.T){
	type mytestStruct struct {
		shellInteractive bool
		userInteractiveConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "true",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file -ti --name=name1 img:1.2.3"},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "false",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3"},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file -ti --name=name1 img:1.2.3"},

		mytestStruct{ shellInteractive: false, userInteractiveConfig: "true",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file -ti --name=name1 img:1.2.3"},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "false",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3"},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3"},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		config.Interactive = v.userInteractiveConfig
		cmd := constructDockerCommand(config, "/tmp/some-env-file", "name1", v.shellInteractive)
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("shellInteractive: %v, userConfig: %v", v.shellInteractive, v.userInteractiveConfig))
	}
}
func Test_ConstructDockerCommand_Command(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3"},
		mytestStruct{ userCommandConfig: "bash",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3 \"bash\""},
		mytestStruct{ userCommandConfig: "bash -c \"echo hello\"",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3 bash -c \"echo hello\""},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		config.DockerRunCommand = v.userCommandConfig
		cmd := constructDockerCommand(config, "/tmp/some-env-file", "name1", false)
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}
func Test_ConstructDockerCommand_DisplayEnvVar(t *testing.T){
	type mytestStruct struct {
		displaySet bool
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ displaySet: true,
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file -v /tmp/.X11-unix:/tmp/.X11-unix --name=name1 img:1.2.3"},
		mytestStruct{ displaySet: false,
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3"},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		if v.displaySet {
			os.Setenv("DISPLAY","123")
		} else {
			setTestEnv()
		}
		cmd := constructDockerCommand(config, "/tmp/some-env-file", "name1", false)
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("displaySet: %v", v.displaySet))
	}
}
func Test_getRunID(t *testing.T) {
	runID := getRunID()
	assert.Contains(t, runID, "dojo-")
	// runID must be lowercase
	lowerCaseRunID := strings.ToLower(runID)
	assert.Equal(t, lowerCaseRunID, runID)
}

func Test_isVariableBlacklisted(t *testing.T) {
	blacklisted := []string{"USER", "PWD", "BASH*"}

	type mytests struct {
		variableName string
		expectedBlacklisted bool
	}
	mytestsObj := []mytests {
		mytests{"PWD", true},
		mytests{"ABC", false},
		mytests{"BASH", true},
		mytests{"BASH111", true},
		mytests{"BAS", false},
		mytests{"BASH*", true},
	}
	for _,v := range mytestsObj {
		actualBlacklisted := isVariableBlacklisted(v.variableName, blacklisted)
		assert.Equal(t, v.expectedBlacklisted, actualBlacklisted, v.variableName)
	}
}

func Test_existsVariableWithDOJOPrefix(t *testing.T) {
	allVariables := []string{"USER=dojo", "BASH_123=123", "DOJO_USER=555", "MYVAR=999"}
	assert.Equal(t, true, existsVariableWithDOJOPrefix("USER", allVariables))
}

func Test_generateVariablesString(t *testing.T) {
	blacklisted := "USER,PWD,BASH*"
	// USER variable is set as USER and DOJO_USER and is blacklisted
	// USER1 variable is set as USER1 and DOJO_USER1 and is not blacklisted
	// BASH_123 variable is blacklisted because of BASH*
	// MYVAR is not blacklisted, is not set with DOJO_ prefix
	// DOJO_VAR1 is not blacklisted, is set with DOJO_ prefix
	// DISPLAY is always set to the same value
	allVariables := []string{"USER=dojo", "BASH_123=123", "DOJO_USER=555", "MYVAR=999", "DOJO_VAR1=11", "USER1=1", "DOJO_USER1=2", "DISPLAY=aaa"}
	genStr := generateVariablesString(blacklisted, allVariables)
	assert.Contains(t, genStr, "DOJO_USER=555")
	assert.Contains(t, genStr, "DOJO_BASH_123=123")
	assert.Contains(t, genStr, "MYVAR=999")
	assert.Contains(t, genStr, "DOJO_VAR1=11")
	assert.Contains(t, genStr, "DOJO_USER1=2")
	assert.Contains(t, genStr, "DISPLAY=unix:0.0")
	assert.NotContains(t, genStr, "DOJO_USER=dojo")
	assert.NotContains(t, genStr, "USER1=1")
	assert.NotContains(t, genStr, "DISPLAY=aaa")
}