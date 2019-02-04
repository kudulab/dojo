package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_smartJoinCommandArgs(t *testing.T){
	var mytests = []struct {
		args    []string
		expectedOutput string
	}{
		{[]string{"cmd"}, "cmd"},
		{[]string{"env | grep MYVAR"}, "\"env | grep MYVAR\""},
		{[]string{"cmd", "cmd and spaces"}, "cmd \"cmd and spaces\""},
		{[]string{"sh", "-c", "echo hello"}, "sh -c \"echo hello\""},
		{[]string{"sh", "-c", "whoami"}, "sh -c whoami"},
		{[]string{"sh", "-c", "env | grep MYVAR"}, "sh -c \"env | grep MYVAR\""},
		// user command: sh -c "env | grep ABC_DEF"
		// or: sh -c 'env | grep ABC_DEF'
		{[]string{"sh", "-c", "env | grep MYVAR"}, "sh -c \"env | grep MYVAR\""},
		// user command: sh -c "docker run -ti \"echo hello\""
		// or: sh -c 'docker run -ti "echo hello"'
		{[]string{"sh", "-c", "docker run -ti \"echo hello\""}, "sh -c \"docker run -ti \\\"echo hello\\\"\""},
		// e.g. entrypoint is /bin/bash, command: -c whoami
		{[]string{"-c", "whoami"}, "-c whoami"},
	}
	for _,v := range mytests {
		outputCmd := smartJoinCommandArgs(v.args)
		assert.Equal(t, outputCmd, v.expectedOutput)
	}
}

func Test_getAbsPathOrPanic(t *testing.T) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var mytests = []struct {
		input          string
		expectedOutput string
	}{
		{"", ""},
		{"/tmp/123", "/tmp/123"},
		{"123", currentDirectory+"/123"},
		{"./123", currentDirectory+"/123"},
	}
	for _,v := range mytests {
		outputCmd := getAbsPathOrPanic(v.input)
		assert.Equal(t, v.expectedOutput, outputCmd)
	}
}

func Test_ensureNoOuterQuotes(t *testing.T) {
	var mytests = []struct {
		input          string
		expectedOutput string
	}{
		{"", ""},
		{"/tmp/123", "/tmp/123"},
		{"\"/tmp/123\"", "/tmp/123"},
		{"'/tmp/123'", "/tmp/123"},
		{"\"/tmp/123", "\"/tmp/123"},
		{"/tmp/123\"", "/tmp/123\""},
		{"'/tmp/123", "'/tmp/123"},
		{"/tmp/123'", "/tmp/123'"},
	}
	for _,v := range mytests {
		outputCmd := ensureNoOuterQuotes(v.input)
		assert.Equal(t, v.expectedOutput, outputCmd, v.input)
	}
}

func Test_getCLIConfig(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	var flagTest = []struct {
		flags    []string
		expectedConfig Config
	}{
		{[]string{"cmd"}, Config{Action:"", ConfigFile:"", Driver:"", Debug:""}},

		{[]string{"cmd", "--config=Dojofile"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:""}},
		{[]string{"cmd", "--config", "Dojofile"}, Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:""}},
		{[]string{"cmd", "-c", "Dojofile"}, Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:""}},
		{[]string{"cmd", "-c=Dojofile"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:""}},

		{[]string{"cmd", "--action", "run"}, Config{Action:"run", ConfigFile:"", Driver:"", Debug:""}},
		{[]string{"cmd", "-a", "run"}, Config{Action:"run", ConfigFile:"", Driver:"", Debug:""}},

		{[]string{"cmd", "--driver", "mydriver"}, Config{Driver:"mydriver"}},
		{[]string{"cmd", "-d", "mydriver"}, Config{Driver:"mydriver"}},

		{[]string{"cmd", "--config=Dojofile", "bash"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:"", RunCommand: "bash"}},
		{[]string{"cmd", "--config=Dojofile", "bash", "bla"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:"", RunCommand: "bash bla"}},
		{[]string{"cmd", "--config=Dojofile", "bash", "-c", "bla"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:"", RunCommand: "bash -c bla"}},
		{[]string{"cmd", "--config=Dojofile", "bash", "-c", "bla1 bla2"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:"", RunCommand: "bash -c \"bla1 bla2\""}},
		{[]string{"cmd", "--config=Dojofile", "bash -c \"bla\""},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:"", RunCommand: "\"bash -c \\\"bla\\\"\""}},
		{[]string{"cmd", "--config=Dojofile", "--", "bash"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:"", RunCommand: "bash"}},
		{[]string{"cmd", "--config=Dojofile", "--", "-c", "bash"},Config{Action:"", ConfigFile:"Dojofile", Driver:"", Debug:"", RunCommand: "-c bash"}},
		{[]string{"cmd", "bash", "--config=Dojofile"},Config{Action:"", ConfigFile:"", Driver:"", Debug:"", RunCommand: "bash --config=Dojofile"}},
		{[]string{"cmd", "--config=Dojofile11", "bash", "--config=Dojofile"},Config{Action:"", ConfigFile:"Dojofile11", Driver:"", Debug:"", RunCommand: "bash --config=Dojofile"}},

		{[]string{"cmd", "--work-dir-outer=/tmp/bla"}, Config{WorkDirOuter:"/tmp/bla"}},
		{[]string{"cmd", "--work-dir-inner=/tmp/bla"}, Config{WorkDirInner:"/tmp/bla"}},
		{[]string{"cmd", "-w=/tmp/bla"}, Config{WorkDirInner:"/tmp/bla"}},
		{[]string{"cmd", "--identity-dir-outer=/tmp/bla"}, Config{IdentityDirOuter: "/tmp/bla"}},
		{[]string{"cmd", "--blacklist=abc,123,ABC_4"}, Config{BlacklistVariables:"abc,123,ABC_4"}},

		{[]string{"cmd", "--action", "run", "-c", "Dojofile"}, Config{Action:"run", ConfigFile:"Dojofile", Driver:"", Debug:""}},
		{[]string{"cmd", "--action", "run", "-c", "Dojofile", "--driver", "mydriver"}, Config{Action:"run", ConfigFile:"Dojofile", Driver:"mydriver", Debug:""}},
		{[]string{"cmd", "--debug=true"}, Config{Action:"", ConfigFile:"", Driver:"", Debug:"true"}},
		{[]string{"cmd", "--debug=false"}, Config{Action:"", ConfigFile:"", Driver:"", Debug:"false"}},
		{[]string{"cmd", "--interactive=false"}, Config{Action:"", ConfigFile:"", Driver:"", Debug:"", Interactive:"false"}},
		{[]string{"cmd", "-i=false"}, Config{Action:"", ConfigFile:"", Driver:"", Debug:"", Interactive:"false"}},
		{[]string{"cmd", "--remove-containers=false"}, Config{RemoveContainers: "false"}},
		{[]string{"cmd", "--rm=false"}, Config{RemoveContainers: "false"}},
	}

	for _, currentTest := range flagTest {
		os.Args = currentTest.flags
		config := getCLIConfig()
		assert.Equal(t, currentTest.expectedConfig.Action, config.Action, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.ConfigFile, config.ConfigFile, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.Driver, config.Driver, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.Debug, config.Debug, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.Interactive, config.Interactive, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.RunCommand, config.RunCommand, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.RemoveContainers, config.RemoveContainers, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.WorkDirOuter, config.WorkDirOuter, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.WorkDirInner, config.WorkDirInner, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.IdentityDirOuter, config.IdentityDirOuter, currentTest.flags)
		assert.Equal(t, currentTest.expectedConfig.BlacklistVariables, config.BlacklistVariables, currentTest.flags)
	}
}
func Test_getCLIConfig_undefinedFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	var flagTest = []struct {
		flags    []string
		expectedAction string
		expectedConfig string
	}{
		{[]string{"cmd", "--notexisting", "ble"}, "", "Dojofile"},
	}

	for _, currentTest := range flagTest {
		os.Args = currentTest.flags

		defer func() {
			expMsg := "flag provided but not defined: -notexisting"
			if r := recover(); !strings.Contains(r.(error).Error(),expMsg) {
				t.Fatalf("Panic message\ngot:  %s\nwant: %s\n", r, expMsg)
			}
		}()

		getCLIConfig()
		t.Fatalf("Expected earlier panic")
	}
}

func Test_getFileConfig(t *testing.T) {
	configFile := "Dojofile-test"
	file, err := os.Create(configFile)
	if err != nil {
		t.Fatal("Cannot create file", err)
	}
	defer file.Close()
	defer os.Remove(configFile)
	fmt.Fprintf(file, "DOJO_DOCKER_IMAGE=docker-registry.example.com/dojo:1.3.2\n")
	fmt.Fprintf(file, "DOJO_DRIVER=somedriver\n")
	fmt.Fprintf(file, "DOJO_DOCKER_OPTIONS=-v /tmp/bla:/home/dojo/bla:ro -e ABC=123\n")
	fmt.Fprintf(file, "DOJO_DOCKER_COMPOSE_FILE=docker-compose.yml\n")
	// absolute path
	fmt.Fprintf(file, "DOJO_WORK_OUTER=/tmp/123\n")
	// relative path
	fmt.Fprintf(file, "DOJO_WORK_INNER=inner\n")
	fmt.Fprintf(file, "DOJO_IDENTITY_OUTER=/tmp/outer\n")
	fmt.Fprintf(file, "DOJO_BLACKLIST_VARIABLES=VAR1,VAR2,ABC\n")
	fmt.Fprintf(file, "DOJO_LOG_LEVEL=info\n")
	fmt.Fprintf(file, "DOJO_PRESERVE_ENV_TO_ALL_CONTAINERS=false\n")

	logger := NewLogger("debug")
	config := getFileConfig(logger, configFile)
	expectedConfig := Config{
		Action:             "",
		DockerImage:        "docker-registry.example.com/dojo:1.3.2",
		Driver:             "somedriver",
		DockerOptions:      "-v /tmp/bla:/home/dojo/bla:ro -e ABC=123",
		PreserveEnvironmentToAllContainers: "false",
		DockerComposeFile:  "docker-compose.yml",
		WorkDirOuter:       "/tmp/123",
		IdentityDirOuter:   "/tmp/outer",
		BlacklistVariables: "VAR1,VAR2,ABC",
		Debug:              "false",
	}

	assert.Equal(t, expectedConfig.Action, config.Action)
	assert.Equal(t, expectedConfig.DockerImage, config.DockerImage)
	assert.Equal(t, expectedConfig.Driver, config.Driver)
	assert.Equal(t, expectedConfig.DockerOptions, config.DockerOptions)
	assert.Equal(t, expectedConfig.DockerComposeFile, config.DockerComposeFile)
	assert.Equal(t, expectedConfig.WorkDirOuter, config.WorkDirOuter)
	// relative path got saved as absolute path
	assert.Contains(t, config.WorkDirInner, "/inner")
	assert.Equal(t, expectedConfig.IdentityDirOuter, config.IdentityDirOuter)
	assert.Equal(t, expectedConfig.BlacklistVariables, config.BlacklistVariables)
	assert.Equal(t, expectedConfig.PreserveEnvironmentToAllContainers, config.PreserveEnvironmentToAllContainers)
	assert.Equal(t, expectedConfig.Debug, config.Debug)
}
func Test_getFileConfig_debug(t *testing.T) {
	configFile := "Dojofile-test1"
	file, err := os.Create(configFile)
	if err != nil {
		t.Fatal("Cannot create file", err)
	}
	defer file.Close()
	defer os.Remove(configFile)
	fmt.Fprintf(file, "DOJO_LOG_LEVEL=debug\n")

	logger := NewLogger("debug")
	config := getFileConfig(logger, configFile)
	expectedConfig := Config{
		Action: "",
		Debug:"true",
	}

	assert.Equal(t, expectedConfig.Debug, config.Debug)
}

func Test_getMergedConfig(t *testing.T){
	config1 := Config{
		Driver: "mydriver",
		Debug: "false",
	}
	config2 := Config{
		Action: "dummy",
		Debug: "true",
		IdentityDirOuter: "/tmp/myhome",
		DockerImage: "img",
	}
	config3 := getDefaultConfig("somefile")
	mergedConfig := getMergedConfig(config1, config2, config3)
	assert.Equal(t, "dummy", mergedConfig.Action)
	assert.Equal(t, "somefile", mergedConfig.ConfigFile)
	assert.Equal(t, "false", mergedConfig.Debug)
	assert.Equal(t, "mydriver", mergedConfig.Driver)
	assert.Equal(t, "true", mergedConfig.RemoveContainers)
	assert.Contains(t, mergedConfig.WorkDirOuter, "/src/dojo")
	assert.Equal(t, "/dojo/work", mergedConfig.WorkDirInner)
	assert.Equal(t, "/tmp/myhome", mergedConfig.IdentityDirOuter)
	assert.Equal(t, "img", mergedConfig.DockerImage)
	assert.Equal(t,
		"BASH*,HOME,USERNAME,USER,LOGNAME,PATH,TERM,SHELL,MAIL,SUDO_*,WINDOWID,SSH_*,SESSION_*,GEM_HOME,GEM_PATH,GEM_ROOT,HOSTNAME,HOSTTYPE,IFS,PPID,PWD,OLDPWD,LC*",
		mergedConfig.BlacklistVariables)
}

func Test_verifyConfig_invalidAction(t *testing.T) {
	config := &Config{
		Action: "dummy",
		Driver: "docker",
		Debug: "true",
	}
	logger := NewLogger("debug")
	err := verifyConfig(logger, config)
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid configuration, unsupported Action: dummy. Supported: run, pull", err.Error())
}

func Test_verifyConfig_invalidDriver(t *testing.T) {
	config := &Config{
		Action: "run",
		Driver: "mydriver",
		Debug: "true",
	}
	logger := NewLogger("debug")
	err := verifyConfig(logger, config)
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid configuration, unsupported Driver: mydriver. Supported: docker, docker-compose", err.Error())
}

func Test_verifyConfig_driverShorthandDC(t *testing.T) {
	dcFile := "/tmp/dojo-Test_verifyConfig_driverShorthandDC.yml"
	config := &Config{
		Action: "run",
		Driver: "dc",
		Debug: "true",
		RemoveContainers: "true",
		DockerImage: "bla",
		DockerComposeFile: dcFile,
		PreserveEnvironmentToAllContainers: "true",
		ExitBehavior: "ignore",
	}
	os.Create(dcFile)
	logger := NewLogger("debug")
	err := verifyConfig(logger, config)
	assert.Nil(t, err)
	assert.Equal(t, "docker-compose", config.Driver)
	os.Remove(dcFile)
}

func Test_mapToConfig(t *testing.T) {
	mymap := make(map[string]string,0)
	mymap["action"] = "run"
	mymap["config"] = "somefile"
	mymap["driver"] = "mydriver"
	mymap["debug"] = "maybe"
	mymap["interactive"] = "meh"
	mymap["removeContainers"] = "true"
	mymap["workDirInner"] = "/tmp/aaa"
	mymap["workDirOuter"] = "/tmp/bbb"
	mymap["identityDirOuter"] = "/tmp/ccc"
	mymap["blacklistVariables"] = "abc"
	mymap["runCommand"] = "whoami"
	mymap["dockerImage"] = "alpine"
	mymap["dockerOptions"] = "-v sth:sth"
	mymap["dockerComposeFile"] = "aaa"
	mymap["dockerComposeOptions"] = "--some-option"
	mymap["preserveEnvironmentToAllContainers"] = "false"
	mymap["exitBehavior"] = "ignore"
	mymap["test"] = "false"
	config := MapToConfig(mymap)
	assert.Equal(t, "mydriver", config.Driver)
	assert.Equal(t, "run", config.Action)

	// assert that all the fields of Config struct are assigned
	v := reflect.ValueOf(config)
	val := reflect.Indirect(reflect.ValueOf(config))
	//values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i).Interface()
		fieldName := val.Type().Field(i).Name
		assert.NotEqual(t, "", fieldValue, fieldName, i)
	}
}