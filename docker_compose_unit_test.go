package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
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
	for _,v := range mytestsObj {
		actualVersion, err := parseDCFileVersion(v.content)
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
	for _,v := range mytestsObj {
		// do not test version, it is tested in other test
		_, err := verifyDCFile(v.content, "filePath.yml")
		if v.expectedErrMsg == "" {
			assert.Equal(t, err, nil)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), v.expectedErrMsg)
		}
	}
}

func Test_generateDCFileContents(t *testing.T) {
	type mytests struct {
		displaySet bool
	}
	mytestsObj := []mytests {
		mytests{true},
		mytests{false},
	}
	for _,v := range mytestsObj {
		config := getTestConfig()
		setTestEnv()
		if v.displaySet {
			os.Setenv("DISPLAY","123")
		} else {
			setTestEnv()
		}
		contents := generateDCFileContents(config, 2.1, "/tmp/env-file.txt")
		assert.Contains(t, contents, "version: '2.1'")
		assert.Contains(t, contents, "  default:")
		assert.Contains(t, contents, "    image: img:1.2.3")
		assert.Contains(t, contents, "    volumes:")
		assert.Contains(t, contents, "      - /tmp/myidentity:/dojo/identity:ro")
		assert.Contains(t, contents, "      - /tmp/bla:/dojo/work")
		assert.Contains(t, contents, "    env_file:")
		assert.Contains(t, contents, "    volumes:")
		assert.Contains(t, contents, "      - /tmp/env-file.txt")
		if v.displaySet {
			assert.Contains(t, contents, "/tmp/.X11-unix")
		} else {
			assert.NotContains(t, contents, "/tmp/.X11-unix")
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
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm --some-opt default bla"},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "false",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm -T --some-opt default bla"},
		mytestStruct{ shellInteractive: true, userInteractiveConfig: "",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm --some-opt default bla"},

		mytestStruct{ shellInteractive: false, userInteractiveConfig: "true",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm --some-opt default bla"},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "false",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm -T --some-opt default bla"},
		mytestStruct{ shellInteractive: false, userInteractiveConfig: "",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm -T --some-opt default bla"},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		config.Interactive = v.userInteractiveConfig
		config.RunCommand = "bla"
		config.DockerComposeOptions = "--some-opt"
		cmd := constructDockerComposeCommandRun(config, "1234", "/tmp/dummy", v.shellInteractive)
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("shellInteractive: %v, userConfig: %v", v.shellInteractive, v.userInteractiveConfig))
	}
}
func Test_ConstructDockerComposeCommandRun_Interactive_NoCommand(t *testing.T){
	setTestEnv()
	config := getTestConfig()
	config.RunCommand = ""
	config.DockerComposeOptions = "--some-opt"

	defer func() {
		if r := recover(); r!= nil {
			assert.Equal(t, "Using driver: docker-compose with empty RunCommand when shell is not interactive is unsupported. It would hang the terminal", r.(string))
		} else {
			t.Fatalf("Expected panic")
		}
	}()

	constructDockerComposeCommandRun(config, "1234", "/tmp/dummy", false)
	t.Fatalf("Expected panic")

}

func Test_ConstructDockerComposeCommandRun(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "bash",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm -T --some-opt default bash"},
		mytestStruct{ userCommandConfig: "bash -c \"echo hello\"",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 run --rm -T --some-opt default bash -c \"echo hello\""},
	}
	setTestEnv()
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		cmd := constructDockerComposeCommandRun(config, "1234", "/tmp/dummy", false)
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
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 stop"},
	}
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		cmd := constructDockerComposeCommandStop(config, "1234", "/tmp/dummy")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}
func Test_ConstructDockerComposeCommandRm(t *testing.T){
	type mytestStruct struct {
		userCommandConfig string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ userCommandConfig: "",
			expOutput: "docker-compose -f docker-compose.yml -f /tmp/dummy -p 1234 rm -f"},
	}
	for _,v := range mytests {
		config := getTestConfig()
		config.RunCommand = v.userCommandConfig
		config.DockerComposeOptions = "--some-opt"
		cmd := constructDockerComposeCommandRm(config, "1234", "/tmp/dummy")
		assert.Equal(t, v.expOutput, cmd, fmt.Sprintf("userCommandConfig: %v", v.userCommandConfig))
	}
}

func Test_getExpDockerNetwork(t *testing.T){
	type mytestStruct struct {
		runID string
		expOutput string
	}
	mytests := []mytestStruct{
		mytestStruct{ runID: "dojo-myproject-2019-01-09_10-39-06-98498093",
			expOutput: "dojomyproject2019010910390698498093_default"},
	}
	for _,v := range mytests {
		expNet := getExpDockerNetwork(v.runID)
		assert.Equal(t, v.expOutput, expNet, v.runID)
	}
}