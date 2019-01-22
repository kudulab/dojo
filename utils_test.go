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

func getTestConfig() Config {
	config := getDefaultConfig("somefile")
	config.DockerImage = "img:1.2.3"
	// set these to some dummy dir, so that tests work also if not run in dojo docker image
	config.WorkDirOuter = "/tmp/bla"
	config.IdentityDirOuter = "/tmp/myidentity"
	return config
}
