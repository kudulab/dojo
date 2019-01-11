package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

func Test_Docker_Run_NoCommand(t *testing.T) {
	config := getTestConfig()
	config.Dryrun = "true"
	config.Driver = "docker"
	runID := "1234"
	exitstatus := handleRun(config, runID)
	assert.Equal(t, 0, exitstatus)
}
func Test_Docker_Run_SimpleCommand(t *testing.T) {
	// set custom Log output target
	var str bytes.Buffer
	log.SetOutput(&str)
	defer log.SetOutput(os.Stdout)

	config := getTestConfig()
	config.Dryrun = "true"
	config.Driver = "docker"
	config.RunCommand = "whoami"
	runID := "1234"
	exitstatus := handleRun(config, runID)
	assert.Equal(t, 0, exitstatus)

	output := strings.TrimSuffix(str.String(), "\n")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "img:1.2.3 \"whoami\"")
}
func Test_Docker_Run_QuotedCommand(t *testing.T) {
	// set custom Log output target
	var str bytes.Buffer
	log.SetOutput(&str)
	defer log.SetOutput(os.Stdout)

	config := getTestConfig()
	config.Dryrun = "true"
	config.Driver = "docker"
	config.RunCommand = "bash -c \"echo hello\""
	runID := "1234"
	exitstatus := handleRun(config, runID)
	assert.Equal(t, 0, exitstatus)

	output := strings.TrimSuffix(str.String(), "\n")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "img:1.2.3 bash -c \"echo hello\"")
}

func Test_Docker_Pull(t *testing.T) {
	config := getTestConfig()
	config.Dryrun = "true"
	config.Driver = "docker"
	exitstatus := handlePull(config)
	assert.Equal(t, 0, exitstatus)
}