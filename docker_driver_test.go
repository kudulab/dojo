package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)


func setTestEnv() {
	os.Unsetenv("DISPLAY")
}

func TestDockerDriver_ConstructDockerRunCmd_Interactive(t *testing.T){
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
		var ss ShellServiceInterface
		if v.shellInteractive {
			ss = MockedShellServiceInteractive{}
		} else {
			ss = MockedShellServiceNotInteractive{}
		}
		d := NewDockerDriver(ss, NewMockedFileService())
		cmd := d.ConstructDockerRunCmd(config, "/tmp/some-env-file", "name1")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("shellInteractive: %v, userConfig: %v", v.shellInteractive, v.userInteractiveConfig))
	}
}
func TestDockerDriver_ConstructDockerRunCmd_Command(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3"},
		mytestStruct{ userCommandConfig: "bash",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3 bash"},
		mytestStruct{ userCommandConfig: "bash -c \"echo hello\"",
			expOutput: "docker run --rm -v /tmp/bla:/dojo/work -v /tmp/myidentity:/dojo/identity:ro --env-file=/tmp/some-env-file --name=name1 img:1.2.3 bash -c \"echo hello\""},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		d := NewDockerDriver(MockedShellServiceNotInteractive{}, NewMockedFileService())
		cmd := d.ConstructDockerRunCmd(config, "/tmp/some-env-file", "name1")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func TestDockerDriver_ConstructDockerRunCmd_DisplayEnvVar(t *testing.T){
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
		d := NewDockerDriver(MockedShellServiceNotInteractive{}, NewMockedFileService())
		cmd := d.ConstructDockerRunCmd(config, "/tmp/some-env-file", "name1")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("displaySet: %v", v.displaySet))
	}
}

func TestDockerDriver_HandleRun_Unit(t *testing.T) {
	fs := NewMockedFileService()
	d := NewDockerDriver(MockedShellServiceNotInteractive{}, fs)
	config := getTestConfig()
	config.RunCommand = ""
	es := d.HandleRun(config, "testrunid", MockedEnvService{})
	assert.Equal(t, 0, es)
	assert.False(t, fileExists("/tmp/dojo-environment-testrunid"))
	assert.Equal(t, 1, len(fs.FilesWrittenTo))
	assert.Equal(t, "ABC=123\n", fs.FilesWrittenTo["/tmp/dojo-environment-testrunid"])
	assert.Equal(t, 2, len(fs.FilesRemovals))
	assert.Equal(t, "/tmp/dojo-environment-testrunid", fs.FilesRemovals[0])
	assert.Equal(t, "/tmp/dojo-environment-testrunid", fs.FilesRemovals[1])
}

func fileExists(filePath string) bool {
	_, err := os.Lstat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(fmt.Sprintf("error when running os.Lstat(%q): %s", filePath, err))
	}
	return true
}

func TestDockerDriver_HandleRun_RealFileService(t *testing.T) {
	d := NewDockerDriver(MockedShellServiceNotInteractive{}, FileService{})
	config := getTestConfig()
	config.WorkDirOuter = "/tmp"
	config.RunCommand = ""
	es := d.HandleRun(config, "testrunid", MockedEnvService{})
	assert.Equal(t, 0, es)
	assert.False(t, fileExists("/tmp/dojo-environment-testrunid"))
}

func TestDockerDriver_HandleRun_RealEnvService(t *testing.T) {
	d := NewDockerDriver(MockedShellServiceNotInteractive{}, NewMockedFileService())
	config := getTestConfig()
	config.RunCommand = ""
	es := d.HandleRun(config, "testrunid", EnvService{})
	assert.Equal(t, 0, es)
	assert.False(t, fileExists("/tmp/dojo-environment-testrunid"))
}

func TestDockerDriver_HandlePull_Unit(t *testing.T) {
	d := NewDockerDriver(MockedShellServiceNotInteractive{}, NewMockedFileService())
	config := getTestConfig()
	es := d.HandlePull(config)
	assert.Equal(t, 0, es)
}