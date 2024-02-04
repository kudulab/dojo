package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLog_UnsupportedLevel(t *testing.T) {
	logger := NewLogger("info")
	defer func() {
		if r := recover(); r != nil {
			expMsg := "Unsupported log level: unsupported"
			if r.(string) != expMsg {
				t.Fatal(fmt.Sprintf("Expected panic message: %v, but was: %v", expMsg, r.(string)))
			}
		}
	}()
	logger.Log("unsupported", "abc")
	t.Fatal("Expected panic, but no panic")
}

func TestLog_Info(t *testing.T) {
	logger := NewLogger("info")
	// set custom Log output target
	var str bytes.Buffer
	logger.SetOutput(&str)

	logger.Log("info", "hello")
	output := strings.TrimSuffix(str.String(), "\n")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "hello")
}

func TestLog_Debug(t *testing.T) {
	logger := NewLogger("debug")
	// set custom Log output target
	var str bytes.Buffer
	logger.SetOutput(&str)

	logger.Log("debug", "my debug msg")
	output := strings.TrimSuffix(str.String(), "\n")
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "my debug msg")
}

func TestLog_MixLevel_DebugUnset(t *testing.T) {
	logger := NewLogger("info")
	// set custom Log output target
	var str bytes.Buffer
	logger.SetOutput(&str)

	logger.Log("info", "hello")
	logger.Log("debug", "my debug msg")
	output := strings.TrimSuffix(str.String(), "\n")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "hello")
	assert.NotContains(t, output, "DEBUG")
	assert.NotContains(t, output, "my debug msg")
}

func TestLog_MixLevel_DebugSet(t *testing.T) {
	logger := NewLogger("debug")
	// set custom Log output target
	var str bytes.Buffer
	logger.SetOutput(&str)

	logger.Log("info", "hello")
	logger.Log("debug", "my debug msg")
	output := strings.TrimSuffix(str.String(), "\n")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "hello")
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "my debug msg")
}

func logSth(logger *Logger) {
	logger.Log("debug", "logging sth")
}

// When you run this test separately, you see output.
func TestLog_MixLevel_ForHuman(t *testing.T) {
	logger := NewLogger("debug")
	// set custom Log output target
	var str bytes.Buffer
	logger.SetOutput(&str)

	logger.Log("info", "hello")
	logger.Log("debug", "my debug msg")
	logSth(logger)
}
func Test_getRunID(t *testing.T) {
	runID := getRunID("false")
	assert.Contains(t, runID, "dojo-")
	// runID must be lowercase
	lowerCaseRunID := strings.ToLower(runID)
	assert.Equal(t, lowerCaseRunID, runID)

	runID = getRunID("true")
	assert.Equal(t, "testdojorunid", runID)
	// runID must be lowercase
	lowerCaseRunID = strings.ToLower(runID)
	assert.Equal(t, lowerCaseRunID, runID)
}

func Test_getRunIDGenerateFromCurrentDir(t *testing.T) {
	runID := getRunIDGenerateFromCurrentDir("mydir")
	assert.True(t, strings.HasPrefix(runID, "dojo-mydir-"))

	runID = getRunIDGenerateFromCurrentDir("mYdIR")
	assert.True(t, strings.HasPrefix(runID, "dojo-mydir-"))
}

func getTestConfig() Config {
	config := getDefaultConfig("somefile")
	config.DockerImage = "img:1.2.3"
	// set these to some dummy dir, so that tests work also if not run in dojo docker image
	config.WorkDirOuter = "/tmp/bla"
	config.IdentityDirOuter = "/tmp/myidentity"
	return config
}

func Test_removeWhiteSpaces(t *testing.T) {
	str := `
aaa

bb
`
	actual := removeWhiteSpaces(str)
	assert.Equal(t, "aaabb", actual)
}

func Test_getContainerInfo(t *testing.T) {
	logger := NewLogger("debug")
	commandsReactions := make(map[string]interface{}, 0)
	fakeOutput := `1234 /name1 running 133`
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' 1234"] =
		[]string{fakeOutput, "", "0"}
	shell := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	info, err := getContainerInfo(shell, "1234")
	assert.Equal(t, "1234", info.ID)
	assert.Equal(t, "name1", info.Name)
	assert.Equal(t, "running", info.Status)
	assert.Equal(t, "133", info.ExitCode)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, info.Exists)
}

func Test_getContainerInfo_NoSuchObject(t *testing.T) {
	logger := NewLogger("debug")
	commandsReactions := make(map[string]interface{}, 0)
	commandsReactions["docker inspect --format='{{.Id}} {{.Name}} {{.State.Status}} {{.State.ExitCode}}' 1234"] =
		[]string{"", "Error: No such object: 1234", "1"}
	shell := NewMockedShellServiceNotInteractive2(logger, commandsReactions)
	info, err := getContainerInfo(shell, "1234")
	assert.Equal(t, "", info.ID)
	assert.Equal(t, "", info.Status)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, info.Exists)
}
