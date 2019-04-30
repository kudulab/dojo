package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	allVariables := []string{"USER=dojo", "BASH_123=123", "DOJO_USER=555", "MYVAR=999", "DOJO_VAR1=11", "USER1=1", "DISPLAY=aaa", "DOJO_USER1=2", "DOJO_WORK_INNER=/my/dir"}
	genStr := generateVariablesString(blacklisted, allVariables)
	assert.Contains(t, genStr, "DOJO_USER=555\n")
	assert.Contains(t, genStr, "DOJO_BASH_123=123\n")
	assert.Contains(t, genStr, "MYVAR=999\n")
	assert.Contains(t, genStr, "DOJO_VAR1=11\n")
	assert.Contains(t, genStr, "DOJO_USER1=2\n")
	assert.Contains(t, genStr, "DISPLAY=unix:0.0\n")
	assert.Contains(t, genStr, "DOJO_WORK_INNER=/my/dir\n")
	assert.NotContains(t, genStr, "DOJO_USER=dojo")
	assert.NotContains(t, genStr, "USER1=1")
	assert.NotContains(t, genStr, "DISPLAY=aaa")
}

func Test_addVariable(t *testing.T) {
	envService := NewEnvService()
	envService.AddVariable("ABC=123")

	assert.Contains(t, envService.Variables, "ABC=123")
}

type MockedEnvService struct {
	Variables []string
}
func NewMockedEnvService() *MockedEnvService {
	return &MockedEnvService{
		Variables: []string{"ABC=123"},
	}
}
func (f MockedEnvService) GetVariables() []string {
	return f.Variables
}
func (f MockedEnvService) IsCurrentUserRoot() bool {
	return false
}
func (f MockedEnvService) AddVariable(keyValue string){
	f.Variables = append(f.Variables, keyValue)
}