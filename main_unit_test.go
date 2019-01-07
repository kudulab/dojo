package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Docker_Run(t *testing.T) {
	config := getTestConfig()
	config.Dryrun = "true"
	config.Driver = "docker"
	exitstatus := handleRun(config)
	assert.Equal(t, 0, exitstatus)
}