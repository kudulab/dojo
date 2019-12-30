package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func Test_parseDCFileVersion(t *testing.T){
	type mytests struct {
		content string
		expectedOutput float64
		expectedErrMsg string
	}
	mytestsObj := []mytests {
		mytests{"version: '4.55'", 4.55, ""},
		mytests{"version: \"4.55\"", 4.55, ""},
		mytests{"", 0, "First line of docker-compose file did not start with: version"},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	for _,v := range mytestsObj {
		actualVersion, err := dc.parseDCFileVersion(v.content)
		assert.Equal(t, v.expectedOutput, actualVersion, v.content)
		if v.expectedErrMsg == "" {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, v.expectedErrMsg, err.Error())
		}
	}
}

func Test_verifyDCFile(t *testing.T) {
	type mytests struct {
		content string
		expectedErrMsg string
	}
	contentsOK := `version: '2'
services:
  default:
    container_name: default
    links:
      - hdind:hdind
`
	contentsNoDefault := `version: '2'
services:
  default123:
    container_name: default
    links:
      - hdind:hdind
`
	contentsInvalidVersion := `version: '3'
services:
  default:
    container_name: default
    links:
      - hdind:hdind
`
	mytestsObj := []mytests {
		mytests{contentsOK, ""},
		mytests{contentsNoDefault, "does not contain: default:"},
		mytests{contentsInvalidVersion, "should contain version number >=2 and <3, current version: 3"},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	for _,v := range mytestsObj {
		// do not test version, it is tested in other test
		_, err := dc.verifyDCFile(v.content, "filePath.yml")
		if v.expectedErrMsg == "" {
			assert.Equal(t, err, nil)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), v.expectedErrMsg)
		}
	}
}

func Test_generateDCFileContentsWithEnv(t *testing.T){
	type mytests struct {
		displaySet bool
	}
	mytestsObj := []mytests {
		mytests{true},
		mytests{false},
	}
	logger := NewLogger("debug")
	expectedServices := []string{"abc", "def", "default"}
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	for _,v := range mytestsObj {
		config := getTestConfig()
		setTestEnv()
		if v.displaySet {
			os.Setenv("DISPLAY","123")
		} else {
			setTestEnv()
		}
		contents := dc.generateDCFileContentsWithEnv(expectedServices, config,
			"/tmp/env-file.txt", "/tmp/env-file-multiline.txt")

		if v.displaySet {
			assert.Equal(t,  `    volumes:
      - /tmp/myidentity:/dojo/identity:ro
      - /tmp/bla:/dojo/work
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
      - /tmp/.X11-unix:/tmp/.X11-unix
    env_file:
      - /tmp/env-file.txt
  abc:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
  def:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
`, contents)
		} else {
			assert.Equal(t,  `    volumes:
      - /tmp/myidentity:/dojo/identity:ro
      - /tmp/bla:/dojo/work
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
    env_file:
      - /tmp/env-file.txt
  abc:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
  def:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
`, contents)
		}
	}

}

func Test_ConstructDockerComposeCommandRun_Interactive(t *testing.T){
	type mytestStruct struct {
		shellInteractive bool
		userInteractiveConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "true",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm --some-opt default bla"},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "false",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bla"},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm --some-opt default bla"},

		mytestStruct{ shellInteractive: false, userInteractiveConfig: "true",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm --some-opt default bla"},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "false",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bla"},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bla"},
	}
	setTestEnv()
	logger := NewLogger("debug")
	for _,v := range mytests {
		config := getTestConfig()
		config.Interactive = v.userInteractiveConfig
		config.RunCommand = "bla"
		config.DockerComposeOptions = "--some-opt"
		config.DockerComposeFile = "/tmp/dummy.yml"
		var ss ShellServiceInterface
		if v.shellInteractive {
			ss = NewMockedShellServiceInteractive(logger)
		} else {
			ss = NewMockedShellServiceNotInteractive(logger)
		}
		dc := NewDockerComposeDriver(ss, NewMockedFileService(logger), logger)
		cmd := dc.ConstructDockerComposeCommandRun(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("shellInteractive: %v, userConfig: %v", v.shellInteractive, v.userInteractiveConfig))
	}
}
func Test_ConstructDockerComposeCommandRun_NotInteractive_NoCommand(t *testing.T){
	setTestEnv()
	config := getTestConfig()
	config.RunCommand = ""
	config.DockerComposeOptions = "--some-opt"
	config.DockerComposeFile = "/tmp/dummy.yml"

	defer func() {
		if r := recover(); r!= nil {
			assert.Equal(t, "Using driver: docker-compose with empty RunCommand when shell is not interactive is unsupported. It would hang the terminal", r.(string))
		} else {
			t.Fatalf("Expected panic")
		}
	}()
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	dc.ConstructDockerComposeCommandRun(config, "1234")
	t.Fatalf("Expected panic")

}

func Test_ConstructDockerComposeCommandRun(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "bash",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bash"},
		mytestStruct{ userCommandConfig: "bash -c \"echo hello\"",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bash -c \"echo hello\""},
	}
	setTestEnv()
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		config.DockerComposeFile = "/tmp/dummy.yml"
		cmd := dc.ConstructDockerComposeCommandRun(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func Test_ConstructDockerComposeCommandStop(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 stop"},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		config.DockerComposeFile = "/tmp/dummy.yml"
		cmd := dc.ConstructDockerComposeCommandStop(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func Test_ConstructDockerComposeCommandDown(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 down"},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger)
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		config.DockerComposeFile = "/tmp/dummy.yml"
		cmd := dc.ConstructDockerComposeCommandDown(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func createDCFile(t *testing.T, dcFilePath string, fs FileServiceInterface)  {
	fs.RemoveFile(dcFilePath, true)
	dcContents := `version: '2'
services:
 default:
   container_name: whatever
`
	fs.WriteToFile(dcFilePath, dcContents, "info")
}
func removeDCFile(dcFilePath string, fs FileServiceInterface)  {
	fs.RemoveFile(dcFilePath, true)
}

func elem_in_array(arr []string, str string) bool {
	for _, a := range arr {
		if strings.Contains(a, str) {
			return true
		}
	}
	return false
}

func TestDockerComposeDriver_HandleRun_Unit(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0" }
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0" }

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	config := getTestConfig()
	config.Driver = "docker-compose"
	config.RunCommand = "bla"
	runID := "1234"
	envService := NewMockedEnvService()
	envService.AddVariable(`MULTI_LINE=one
two
three`)
	exitstatus := driver.HandleRun(config, runID, envService)
	assert.Equal(t, 0, exitstatus)
	assert.Equal(t, 3, len(fs.FilesWrittenTo))
	assert.Equal(t, "ABC=123\n", fs.FilesWrittenTo["/tmp/dojo-environment-1234"])
	assert.Equal(t, "export MULTI_LINE=$(echo b25lCnR3bwp0aHJlZQ== | base64 -d)\n", fs.FilesWrittenTo["/tmp/dojo-environment-multiline-1234"])
	assert.Contains(t, fs.FilesWrittenTo["docker-compose.yml.dojo"], "version: '2.2'")

	exitstatus = driver.CleanAfterRun(config, runID)
	assert.Equal(t, 0, exitstatus)
	assert.Equal(t, 5, len(fs.FilesRemovals))
	assert.Equal(t, "/tmp/dojo-environment-1234", fs.FilesRemovals[0])
	assert.Equal(t, "/tmp/dojo-environment-multiline-1234", fs.FilesRemovals[1])
	assert.Equal(t, "docker-compose.yml.dojo", fs.FilesRemovals[2])
	assert.False(t, fileExists("/tmp/dojo-environment-1234"))
}

func TestDockerComposeDriver_HandleRun_Unit_PrintLogsFailure(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0" }
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0" }

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	config := getTestConfig()
	config.Driver = "docker-compose"
	config.RunCommand = "bla"
	config.PrintLogs = "failure"
	runID := "1234"
	envService := NewMockedEnvService()
	exitstatus := driver.HandleRun(config, runID, envService)
	assert.Equal(t, 0, exitstatus)
	assert.False(t, elem_in_array(shellS.CommandsRun, "docker logs"))

	exitstatus = driver.CleanAfterRun(config, runID)
	assert.Equal(t, 0, exitstatus)
}

func TestDockerComposeDriver_HandleRun_Unit_PrintLogsAlways(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0" }
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0" }

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	config := getTestConfig()
	config.Driver = "docker-compose"
	config.RunCommand = "bla"
	config.PrintLogs = "always"
	runID := "1234"
	envService := NewMockedEnvService()
	exitstatus := driver.HandleRun(config, runID, envService)
	assert.Equal(t, 0, exitstatus)
	assert.True(t, elem_in_array(shellS.CommandsRun, "docker logs"))

	exitstatus = driver.CleanAfterRun(config, runID)
	assert.Equal(t, 0, exitstatus)
}

func TestDockerComposeDriver_HandleRun_RealFileService(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	dcFilePath := "test-docker-compose.yml"
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f test-docker-compose.yml -f test-docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0" }
	commandsReactions["docker-compose -f test-docker-compose.yml -f test-docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0" }

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	createDCFile(t, dcFilePath, fs)
	config := getTestConfig()
	config.Driver = "docker-compose"
	config.DockerComposeFile = dcFilePath
	config.WorkDirOuter = "/tmp"
	config.RunCommand = "bla"
	runID := "1234"
	exitstatus := driver.HandleRun(config, runID, NewMockedEnvService())
	assert.Equal(t, 0, exitstatus)
	exitstatus = driver.CleanAfterRun(config, runID)
	assert.Equal(t, 0, exitstatus)
	removeDCFile(dcFilePath, fs)
	removeDCFile(dcFilePath+".dojo", fs)
	fs.RemoveFile("/tmp/dojo-environment-1234", true)
}

func getFakeDockerComposePSStdout() string {
	return `Name                        Command               State   Ports
------------------------------------------------------------------------
edudocker_abc_1           /bin/sh -c while true; do  ...   Up
edudocker_def_1           /bin/sh -c while true; do  ...   Up
edudocker_default_run_1   sh -c echo 'will sleep' && ...   Up
`
}

func TestDockerComposeDriver_HandleRun_RealEnvService(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0" }
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0" }

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	config := getTestConfig()
	config.Driver = "docker-compose"
	config.WorkDirOuter = "/tmp"
	config.RunCommand = "bla"
	runID := "1234"
	exitstatus := driver.HandleRun(config, runID, NewEnvService())
	assert.Equal(t, 0, exitstatus)
	exitstatus = driver.CleanAfterRun(config, runID)
	assert.Equal(t, 0, exitstatus)
}

func Test_getDCContainersNames(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0" }

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	names := driver.getDCContainersNames(getTestConfig(), "1234")
	assert.Equal(t, 3, len(names))
	assert.Equal(t, "edudocker_abc_1", names[0])
	assert.Equal(t, "edudocker_def_1", names[1])
	assert.Equal(t, "edudocker_default_run_1", names[2])
}

func Test_getDefaultContainerID(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)

	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker inspect --format='{{.Id}} {{.State.Status}}' edudocker_default_run_1"] =
		[]string{"dummy-id running", "", "0" }
	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	names := []string{"edudocker_abc_1", "edudocker_def_1", "edudocker_default_run_1"}
	id := driver.getDefaultContainerID(names)
	assert.Equal(t, "dummy-id", id)
}

func Test_getDefaultContainerID_notCreated(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)

	shellS := NewMockedShellServiceNotInteractive(logger)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	defer func() {
		r := recover()
		assert.Contains(t, r.(error).Error(), "default container not found. Were the containers created?")
	}()

	names := []string{}
	id := driver.getDefaultContainerID(names)
	assert.Equal(t, "", id)
	t.Fatal("Expected panic")
}

func Test_checkContainerIsRunning(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)

	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker inspect --format='{{.Id}} {{.State.Status}}' id1"] =	[]string{"id1 running", "", "0" }
	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	running := driver.checkContainerIsRunning("id1")
	assert.Equal(t, true, running)
}

func Test_waitForContainersToBeRunning(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0" }
	commandsReactions["docker inspect --format='{{.Id}} {{.State.Status}}' edudocker_abc_1"] =
		[]string{"id1 running", "", "0" }
	commandsReactions["docker inspect --format='{{.Id}} {{.State.Status}}' edudocker_def_1"] =
		[]string{"id2 running", "", "0" }
	commandsReactions["docker inspect --format='{{.Id}} {{.State.Status}}' edudocker_default_run_1"] =
		[]string{"id3 running", "", "0" }
	fakeContainers := `abc
cde
efd
`
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{fakeContainers, "", "0" }
	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger)

	ids := driver.waitForContainersToBeRunning(getTestConfig(), "1234", 3)
	assert.Equal(t, []string{"edudocker_abc_1", "edudocker_def_1", "edudocker_default_run_1"}, ids)
}

func Test_getExpectedContainers(t *testing.T) {
	type mytests struct {
		fakeOutput string
		fakeExitStatus int
		expectedNames []string
		expectedError string
	}
	output1 := `abc
default
`
	output2 := `abc
default`
	mytestsObj := []mytests {
		mytests{output1,0, []string{"abc", "default"}, ""},
		mytests{output2, 0, []string{"abc", "default"}, ""},
		mytests{"", 1, []string{}, "Exit status: 1"},
	}

	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)

	for _, tt := range mytestsObj {
		commandsReactions := make(map[string]interface{}, 0)
		commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
			[]string{tt.fakeOutput, "", fmt.Sprintf("%v",tt.fakeExitStatus) }
		shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
		driver := NewDockerComposeDriver(shellS, fs, logger)

		if tt.expectedError == "" {
			containers := driver.getExpectedContainers(getTestConfig(), "1234")
			assert.Equal(t, tt.expectedNames, containers)
		} else {
			defer func() {
				if r := recover(); r!= nil {
					errMsg := r.(error).Error()
					assert.Contains(t, errMsg, tt.expectedError)
				} else {
					t.Fatalf("Expected panic")
				}
			}()
			driver.getExpectedContainers(getTestConfig(), "1234")
			t.Fatalf("Expected panic")
		}
	}
}