package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func Test_parseDCFileVersion(t *testing.T) {
	type mytests struct {
		content        string
		expectedOutput float64
		expectedErrMsg string
	}
	mytestsObj := []mytests{
		mytests{"version: '4.55'", 4.55, ""},
		mytests{"version: \"4.55\"", 4.55, ""},
		mytests{"", -1, ""},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger, "")
	for _, v := range mytestsObj {
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
		content        string
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
	mytestsObj := []mytests{
		mytests{contentsOK, ""},
		mytests{contentsNoDefault, "does not contain: default:"},
		mytests{contentsInvalidVersion, "should contain version number >=2 and <3, current version: 3"},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger, "")
	for _, v := range mytestsObj {
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

func Test_generateDCFileContentsWithEnv(t *testing.T) {
	type mytests struct {
		displaySet bool
	}
	mytestsObj := []mytests{
		mytests{true},
		mytests{false},
	}
	logger := NewLogger("debug")
	expectedServices := []string{"abc", "def", "default"}
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger, "")
	for _, v := range mytestsObj {
		config := getTestConfig()
		setTestEnv()
		if v.displaySet {
			os.Setenv("DISPLAY", "123")
		} else {
			setTestEnv()
		}
		contents := dc.generateDCFileContentsWithEnv(expectedServices, config,
			"/tmp/env-file.txt", "/tmp/env-file-multiline.txt", "/tmp/env-file-bash-functions.txt")

		if v.displaySet {
			assert.Equal(t, `    volumes:
      - /tmp/myidentity:/dojo/identity:ro
      - /tmp/bla:/dojo/work
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
      - /tmp/env-file-bash-functions.txt:/etc/dojo.d/variables/01-bash-functions.sh
      - /tmp/.X11-unix:/tmp/.X11-unix
    env_file:
      - /tmp/env-file.txt
  abc:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
      - /tmp/env-file-bash-functions.txt:/etc/dojo.d/variables/01-bash-functions.sh
  def:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
      - /tmp/env-file-bash-functions.txt:/etc/dojo.d/variables/01-bash-functions.sh
`, contents)
		} else {
			assert.Equal(t, `    volumes:
      - /tmp/myidentity:/dojo/identity:ro
      - /tmp/bla:/dojo/work
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
      - /tmp/env-file-bash-functions.txt:/etc/dojo.d/variables/01-bash-functions.sh
    env_file:
      - /tmp/env-file.txt
  abc:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
      - /tmp/env-file-bash-functions.txt:/etc/dojo.d/variables/01-bash-functions.sh
  def:
    env_file:
      - /tmp/env-file.txt
    volumes:
      - /tmp/env-file-multiline.txt:/etc/dojo.d/variables/00-multiline-vars.sh
      - /tmp/env-file-bash-functions.txt:/etc/dojo.d/variables/01-bash-functions.sh
`, contents)
		}
	}

}

func Test_ConstructDockerComposeCommandRun_Interactive(t *testing.T) {
	type mytestStruct struct {
		shellInteractive      bool
		userInteractiveConfig string
		expOutput             string
	}
	mytests := []mytestStruct{
		mytestStruct{shellInteractive: true, userInteractiveConfig: "true",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm --some-opt default bla"},
		mytestStruct{shellInteractive: true, userInteractiveConfig: "false",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bla"},
		mytestStruct{shellInteractive: true, userInteractiveConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm --some-opt default bla"},

		mytestStruct{shellInteractive: false, userInteractiveConfig: "true",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm --some-opt default bla"},
		mytestStruct{shellInteractive: false, userInteractiveConfig: "false",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bla"},
		mytestStruct{shellInteractive: false, userInteractiveConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bla"},
	}
	setTestEnv()
	logger := NewLogger("debug")
	for _, v := range mytests {
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
		dc := NewDockerComposeDriver(ss, NewMockedFileService(logger), logger, "")
		cmd := dc.ConstructDockerComposeCommandRun(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("shellInteractive: %v, userConfig: %v", v.shellInteractive, v.userInteractiveConfig))
	}
}
func Test_ConstructDockerComposeCommandRun_NotInteractive_NoCommand(t *testing.T) {
	setTestEnv()
	config := getTestConfig()
	config.RunCommand = ""
	config.DockerComposeOptions = "--some-opt"
	config.DockerComposeFile = "/tmp/dummy.yml"

	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "Using driver: docker-compose with empty RunCommand when shell is not interactive is unsupported. It would hang the terminal", r.(string))
		} else {
			t.Fatalf("Expected panic")
		}
	}()
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger, "")
	dc.ConstructDockerComposeCommandRun(config, "1234")
	t.Fatalf("Expected panic")

}

func Test_ConstructDockerComposeCommandRun(t *testing.T) {
	type mytestStruct struct {
		userCommandConfig string
		expOutput         string
	}
	mytests := []mytestStruct{
		mytestStruct{userCommandConfig: "bash",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bash"},
		mytestStruct{userCommandConfig: "bash -c \"echo hello\"",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 run --rm -T --some-opt default bash -c \"echo hello\""},
	}
	setTestEnv()
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger, "")
	for _, v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		config.DockerComposeFile = "/tmp/dummy.yml"
		cmd := dc.ConstructDockerComposeCommandRun(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func Test_ConstructDockerComposeCommandStop(t *testing.T) {
	type mytestStruct struct {
		userCommandConfig string
		expOutput         string
	}
	mytests := []mytestStruct{
		mytestStruct{userCommandConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 stop"},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger, "")
	for _, v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		config.DockerComposeFile = "/tmp/dummy.yml"
		cmd := dc.ConstructDockerComposeCommandStop(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func Test_ConstructDockerComposeCommandDown(t *testing.T) {
	type mytestStruct struct {
		userCommandConfig string
		expOutput         string
	}
	mytests := []mytestStruct{
		mytestStruct{userCommandConfig: "",
			expOutput: "docker-compose -f /tmp/dummy.yml -f /tmp/dummy.yml.dojo -p 1234 down"},
	}
	logger := NewLogger("debug")
	dc := NewDockerComposeDriver(NewMockedShellServiceNotInteractive(logger), NewMockedFileService(logger), logger, "")
	for _, v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		config.DockerComposeFile = "/tmp/dummy.yml"
		cmd := dc.ConstructDockerComposeCommandDown(config, "1234")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func createDCFile(t *testing.T, dcFilePath string, fs FileServiceInterface) {
	fs.RemoveFile(dcFilePath, true)
	dcContents := `version: '2'
services:
 default:
   container_name: whatever
`
	fs.WriteToFile(dcFilePath, dcContents, "info")
}
func removeDCFile(dcFilePath string, fs FileServiceInterface) {
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
		[]string{fakePSOutput, "", "0"}
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_abc_1"] =
		[]string{"container1 name1 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_def_1"] =
		[]string{"container1 name2 running 0", "", "0"}

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

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
	assert.Equal(t, 4, len(fs.FilesWrittenTo))
	assert.Equal(t, "ABC=123\n", fs.FilesWrittenTo["/tmp/dojo-environment-1234"])
	assert.Equal(t, "export MULTI_LINE=$(echo b25lCnR3bwp0aHJlZQ== | base64 -d)\n", fs.FilesWrittenTo["/tmp/dojo-environment-multiline-1234"])
	assert.Contains(t, fs.FilesWrittenTo["docker-compose.yml.dojo"], "version: '2.2'")

	exitstatus = driver.CleanAfterRun(config, runID)
	assert.Equal(t, 0, exitstatus)
	assert.Equal(t, 7, len(fs.FilesRemovals))
	assert.Equal(t, "/tmp/dojo-environment-1234", fs.FilesRemovals[1])
	assert.Equal(t, "/tmp/dojo-environment-multiline-1234", fs.FilesRemovals[2])
	assert.Equal(t, "/tmp/dojo-environment-bash-functions-1234", fs.FilesRemovals[0])
	assert.Equal(t, "docker-compose.yml.dojo", fs.FilesRemovals[3])
	assert.False(t, fileExists("/tmp/dojo-environment-1234"))
}

func TestDockerComposeDriver_HandleRun_Unit_PrintLogsFailure(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0"}
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_abc_1"] =
		[]string{"container1 name1 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_def_1"] =
		[]string{"container1 name1 running 0", "", "0"}

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

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

func TestDockerComposeDriver_HandleRun_Unit_PrintLogsAlways_TargetConsole(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0"}
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_abc_1"] = []string{"some_hash name1 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_def_1"] = []string{"some_hash name2 running 127", "", "0"}

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

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
func TestDockerComposeDriver_HandleRun_Unit_PrintLogsAlways_TargetFile(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0"}
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_abc_1"] = []string{"some_hash name1 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_def_1"] = []string{"some_hash name2 running 127", "", "0"}
	commandsReactions["docker logs name1"] = []string{"some-output", "", "0"}
	commandsReactions["docker logs name2"] = []string{"some-output2", "", "0"}

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

	config := getTestConfig()
	config.Driver = "docker-compose"
	config.RunCommand = "bla"
	config.PrintLogs = "always"
	config.PrintLogsTarget = "file"
	runID := "1234"
	envService := NewMockedEnvService()
	exitstatus := driver.HandleRun(config, runID, envService)
	assert.Equal(t, 0, exitstatus)
	assert.True(t, elem_in_array(shellS.CommandsRun, "docker logs"))
	assert.Contains(t, fs.FilesWrittenTo["dojo-logs-name1-1234.txt"], "stderr:\nstdout:\nsome-output")
	assert.Contains(t, fs.FilesWrittenTo["dojo-logs-name2-1234.txt"], "stderr:\nstdout:\nsome-output2")

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
		[]string{fakePSOutput, "", "0"}
	commandsReactions["docker-compose -f test-docker-compose.yml -f test-docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_abc_1"] = []string{"some_hash name1 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_def_1"] = []string{"some_hash name2 running 127", "", "0"}

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

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
		[]string{fakePSOutput, "", "0"}
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{"container1", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_abc_1"] = []string{"some_hash name1 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_def_1"] = []string{"some_hash name2 running 127", "", "0"}

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

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
		[]string{fakePSOutput, "", "0"}

	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

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
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_default_run_1"] =
		[]string{"dummy-id name1 running 000", "", "0"}
	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

	names := []string{"edudocker_abc_1", "edudocker_def_1", "edudocker_default_run_1"}
	id := driver.getDefaultContainerID(names)
	assert.Equal(t, "dummy-id", id)
}

func Test_getDefaultContainerID_notCreated(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)

	shellS := NewMockedShellServiceNotInteractive(logger)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

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
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' id1"] = []string{"id1 name1 running 0", "", "0"}
	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

	running := driver.checkContainerIsRunning("id1")
	assert.Equal(t, true, running)
}

func Test_waitForContainersToBeRunning(t *testing.T) {
	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	fakePSOutput := getFakeDockerComposePSStdout()
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 ps"] =
		[]string{fakePSOutput, "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_abc_1"] =
		[]string{"id1 name1 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_def_1"] =
		[]string{"id2 name2 running 0", "", "0"}
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' edudocker_default_run_1"] =
		[]string{"id3 name3 running 0", "", "0"}
	fakeContainers := `abc
cde
efd
`
	commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
		[]string{fakeContainers, "", "0"}
	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")

	ids := driver.waitForContainersToBeRunning(getTestConfig(), "1234", 3)
	assert.Equal(t, []string{"edudocker_abc_1", "edudocker_def_1", "edudocker_default_run_1"}, ids)
}

func Test_getExpectedContainers(t *testing.T) {
	type mytests struct {
		fakeOutput     string
		fakeExitStatus int
		expectedNames  []string
		expectedError  string
	}
	output1 := `abc
default
`
	output2 := `abc
default`
	mytestsObj := []mytests{
		mytests{output1, 0, []string{"abc", "default"}, ""},
		mytests{output2, 0, []string{"abc", "default"}, ""},
		mytests{"", 1, []string{}, "Exit status: 1"},
	}

	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)

	for _, tt := range mytestsObj {
		commandsReactions := make(map[string]interface{}, 0)
		commandsReactions["docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p 1234 config --services"] =
			[]string{tt.fakeOutput, "", fmt.Sprintf("%v", tt.fakeExitStatus)}
		shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
		driver := NewDockerComposeDriver(shellS, fs, logger, "")

		if tt.expectedError == "" {
			containers := driver.getExpectedContainers(getTestConfig(), "1234")
			assert.Equal(t, tt.expectedNames, containers)
		} else {
			defer func() {
				if r := recover(); r != nil {
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

func Test_checkIfAnyContainerFailed_allSucceeded(t *testing.T) {
	nonDefContInfos := make([]*ContainerInfo, 0)
	cont1Info := &ContainerInfo{
		ExitCode: "0",
	}
	nonDefContInfos = append(nonDefContInfos, cont1Info)
	cont2Info := &ContainerInfo{
		ExitCode: "0",
	}
	nonDefContInfos = append(nonDefContInfos, cont2Info)
	anyContFailed := checkIfAnyContainerFailed(nonDefContInfos, 0)
	assert.False(t, anyContFailed)
}
func Test_checkIfAnyContainerFailed_defaultFailed(t *testing.T) {
	nonDefContInfos := make([]*ContainerInfo, 0)
	cont1Info := &ContainerInfo{
		ExitCode: "0",
	}
	nonDefContInfos = append(nonDefContInfos, cont1Info)
	cont2Info := &ContainerInfo{
		ExitCode: "0",
	}
	nonDefContInfos = append(nonDefContInfos, cont2Info)
	anyContFailed := checkIfAnyContainerFailed(nonDefContInfos, 3)
	assert.True(t, anyContFailed)
}
func Test_checkIfAnyContainerFailed_nonDefFailed(t *testing.T) {
	nonDefContInfos := make([]*ContainerInfo, 0)
	cont1Info := &ContainerInfo{
		ExitCode: "0",
	}
	nonDefContInfos = append(nonDefContInfos, cont1Info)
	cont2Info := &ContainerInfo{
		ExitCode: "144",
	}
	nonDefContInfos = append(nonDefContInfos, cont2Info)
	anyContFailed := checkIfAnyContainerFailed(nonDefContInfos, 0)
	assert.True(t, anyContFailed)
}

func Test_getNonDefaultContainersLogs(t *testing.T) {
	nonDefContInfos := make([]*ContainerInfo, 0)
	cont1Info := &ContainerInfo{
		Name: "name1",
	}
	nonDefContInfos = append(nonDefContInfos, cont1Info)
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker logs name1"] =
		[]string{"123", "", "0"}

	logger := NewLogger("debug")
	fs := NewMockedFileService(logger)
	shellS := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	driver := NewDockerComposeDriver(shellS, fs, logger, "")
	driver.getNonDefaultContainersLogs(nonDefContInfos)
	assert.Equal(t, nonDefContInfos[0].Logs, "stderr:\nstdout:\n123")
}

func Test_parseDCPSOutPut_DCVersion2_whenOutputNonEmpty(t *testing.T) {
	dc_ps_output := "{\"Command\":\"\\\"/bin/sh -c 'while t…\\\"\",\"CreatedAt\":\"2024-02-03 21:03:46 +0000 UTC\",\"ExitCode\":0,\"Health\":\"\",\"ID\":\"2d5c5b0343d0\",\"Image\":\"alpine:3.19\",\"Labels\":\"com.docker.compose.depends_on=,com.docker.compose.image=sha256:05455a08881ea9cf0e752bc48e61bbd71a34c029bb13df01e40e3e70e0d007bd,com.docker.compose.version=2.24.5,com.docker.compose.service=abc,com.docker.compose.config-hash=270e27422cb1e6a4c1713ae22a3ffca0e8aa50ec0f06fe493fa4f83a17bd29e9,com.docker.compose.container-number=1,com.docker.compose.oneoff=False,com.docker.compose.project=testdojorunid,com.docker.compose.project.config_files=/dojo/work/test/test-files/itest-dc.yaml,/dojo/work/test/test-files/itest-dc.yaml.dojo,com.docker.compose.project.working_dir=/dojo/work/test/test-files\",\"LocalVolumes\":\"0\",\"Mounts\":\"/tmp/test-dojo…,/tmp/test-dojo…\",\"Name\":\"testdojorunid-abc-1\",\"Names\":\"testdojorunid-abc-1\",\"Networks\":\"testdojorunid_default\",\"Ports\":\"\",\"Project\":\"testdojorunid\",\"Publishers\":null,\"RunningFor\":\"3 seconds ago\",\"Service\":\"abc\",\"Size\":\"0B\",\"State\":\"running\",\"Status\":\"Up 2 seconds\"}\n{\"Command\":\"\\\"/bin/sh -c 'while t…\\\"\",\"CreatedAt\":\"2024-02-03 21:03:46 +0000 UTC\",\"ExitCode\":0,\"Health\":\"\",\"ID\":\"b2ed210567c3\",\"Image\":\"alpine:3.19\",\"Labels\":\"com.docker.compose.project=testdojorunid,com.docker.compose.project.config_files=/dojo/work/test/test-files/itest-dc.yaml,/dojo/work/test/test-files/itest-dc.yaml.dojo,com.docker.compose.project.working_dir=/dojo/work/test/test-files,com.docker.compose.depends_on=,com.docker.compose.container-number=1,com.docker.compose.image=sha256:05455a08881ea9cf0e752bc48e61bbd71a34c029bb13df01e40e3e70e0d007bd,com.docker.compose.oneoff=False,com.docker.compose.service=def,com.docker.compose.version=2.24.5,com.docker.compose.config-hash=270e27422cb1e6a4c1713ae22a3ffca0e8aa50ec0f06fe493fa4f83a17bd29e9\",\"LocalVolumes\":\"0\",\"Mounts\":\"/tmp/test-dojo…,/tmp/test-dojo…\",\"Name\":\"testdojorunid-def-1\",\"Names\":\"testdojorunid-def-1\",\"Networks\":\"testdojorunid_default\",\"Ports\":\"\",\"Project\":\"testdojorunid\",\"Publishers\":null,\"RunningFor\":\"3 seconds ago\",\"Service\":\"def\",\"Size\":\"0B\",\"State\":\"running\",\"Status\":\"Up 2 seconds\"}\n{\"Command\":\"\\\"sh -c 'sleep 10'\\\"\",\"CreatedAt\":\"2024-02-03 21:03:47 +0000 UTC\",\"ExitCode\":0,\"Health\":\"\",\"ID\":\"af4817fede41\",\"Image\":\"alpine:3.15\",\"Labels\":\"com.docker.compose.version=2.24.5,com.docker.compose.container-number=1,com.docker.compose.depends_on=abc:service_started:true,def:service_started:true,com.docker.compose.oneoff=True,com.docker.compose.project=testdojorunid,com.docker.compose.slug=742bcbb0e4bc05b21928a8d17be4ea9bb12a6775fd40692dd59c74a460279eb8,com.docker.compose.config-hash=462afacb4521d13580c2096c7b00b98970f07fe841e408c4c5a95a4a46839eaa,com.docker.compose.image=sha256:32b91e3161c8fc2e3baf2732a594305ca5093c82ff4e0c9f6ebbd2a879468e1d,com.docker.compose.project.config_files=/dojo/work/test/test-files/itest-dc.yaml,/dojo/work/test/test-files/itest-dc.yaml.dojo,com.docker.compose.project.working_dir=/dojo/work/test/test-files,com.docker.compose.service=default\",\"LocalVolumes\":\"0\",\"Mounts\":\"/home/dojo,/dojo/work,/tmp/test-dojo…,/tmp/test-dojo…,/tmp/.X11-unix,/tmp/dojo-ites…\",\"Name\":\"testdojorunid-default-run-742bcbb0e4bc\",\"Names\":\"testdojorunid-default-run-742bcbb0e4bc\",\"Networks\":\"testdojorunid_default\",\"Ports\":\"\",\"Project\":\"testdojorunid\",\"Publishers\":null,\"RunningFor\":\"2 seconds ago\",\"Service\":\"default\",\"Size\":\"0B\",\"State\":\"running\",\"Status\":\"Up 1 second\"}\n"
	output_as_json, err := ParseDCPSOutPut_DCVersion2(dc_ps_output)
	assert.Equal(t, "", err)
	assert.Equal(t, 3, len(output_as_json))
	assert.Equal(t, "2024-02-03 21:03:46 +0000 UTC", output_as_json[0].CreatedAt)
	assert.Equal(t, "Up 1 second", output_as_json[2].Status)
}

func Test_parseDCPSOutPut_DCVersion2_whenOutputEmpty(t *testing.T) {
	dc_ps_output := ""
	output_as_json, err := ParseDCPSOutPut_DCVersion2(dc_ps_output)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, len(output_as_json))
}

func Test_parseDCPSOutPut_DCVersion2_whenOutputNonEmptyButInvalid(t *testing.T) {
	dc_ps_output := "{\"Command123\":\"\\\"/bin/sh -c 'while t…\\\"\"}"
	output_as_json, err := ParseDCPSOutPut_DCVersion2(dc_ps_output)
	assert.Contains(t, err, "State was an empty string")
	assert.Equal(t, 0, len(output_as_json))
}

func Test_isDCVersionLaterThan2_v2WithoutV(t *testing.T) {
	assert.Equal(t, true, isDCVersionLaterThan2("2.24.5"))
}

func Test_isDCVersionLaterThan2_v2WithV(t *testing.T) {
	assert.Equal(t, true, isDCVersionLaterThan2("v2.24.5"))
}

func Test_isDCVersionLaterThan2_v1WithoutV(t *testing.T) {
	assert.Equal(t, false, isDCVersionLaterThan2("1.24.5"))
}

func Test_isDCVersionLaterThan2_v1WithV(t *testing.T) {
	assert.Equal(t, false, isDCVersionLaterThan2("v1.24.5"))
}

func Test_isDCVersionLaterThan2_empty(t *testing.T) {
	assert.Equal(t, false, isDCVersionLaterThan2(""))
}
