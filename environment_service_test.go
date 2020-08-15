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

func Test_filterBlacklistedVariables(t *testing.T) {
	blacklisted := "USER,PWD,BASH*"
	// USER variable is set as USER and DOJO_USER and is blacklisted
	// USER1 variable is set as USER1 and DOJO_USER1 and is not blacklisted
	// BASH_123 variable is blacklisted because of BASH*
	// MYVAR is not blacklisted, is not set with DOJO_ prefix
	// DOJO_VAR1 is not blacklisted, is set with DOJO_ prefix
	// DISPLAY is always set to the same value
	allVariables := []string{
		"USER=dojo", "BASH_123=123", "DOJO_USER=555", "MYVAR=999",
		"DOJO_VAR1=11", "USER1=1", "DISPLAY=aaa", "DOJO_USER1=2",
		"DOJO_WORK_INNER=/my/dir", `MULTI_LINE=one
two`}
	filteredEnvVariables := filterBlacklistedVariables(blacklisted, allVariables)
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"DOJO_USER", "555", false, false})
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"DOJO_BASH_123", "123", false, false})
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"MYVAR", "999", false, false})
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"DOJO_VAR1", "11", false, false})
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"DOJO_USER1", "2", false, false})
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"DISPLAY", "unix:0.0", false, false})
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"DOJO_WORK_INNER", "/my/dir", false, false})
	assert.Contains(t, filteredEnvVariables, EnvironmentVariable{"MULTI_LINE", `one
two`, true, false})
	assert.NotContains(t, filteredEnvVariables, EnvironmentVariable{"DOJO_USER", "dojo", false, false})
	assert.NotContains(t, filteredEnvVariables, EnvironmentVariable{"USER1", "1", false, false})
	assert.NotContains(t, filteredEnvVariables, EnvironmentVariable{"DISPLAY", "aaa", false, false})
}

func Test_singleLineVariablesToString(t *testing.T) {
	allVariables := []EnvironmentVariable{
		EnvironmentVariable{"DOJO_USER", "555", false, false},
		EnvironmentVariable{"DOJO_WORK_INNER", "/my/dir", false, false},
		EnvironmentVariable{"MULTI_LINE", `one
two`, true, false},
		EnvironmentVariable{"MYVAR", "999", false, false},
	}
	genStr := singleLineVariablesToString(allVariables)
	assert.Contains(t, genStr, "DOJO_USER=555\n")
	assert.Contains(t, genStr, "MYVAR=999\n")
	assert.Contains(t, genStr, "DOJO_WORK_INNER=/my/dir\n")
	assert.NotContains(t, genStr, "MULTI_LINE")
}

func Test_checkIfBashFunc(t *testing.T) {
	assert.True(t, checkIfBashFunc("() { echo hello }", "BASH_FUNC_abc_%%"))
	assert.False(t, checkIfBashFunc("text", "BASH_FUNC_abc_%%"))
	assert.False(t, checkIfBashFunc("() { echo hello }", "abc_%%"))
}

func Test_EnvVar_encryptValue(t *testing.T) {
	e := EnvironmentVariable{"DOJO_USER", "555", false, false}
	str := e.encryptValue()
	assert.Equal(t, "NTU1", str)
}

func Test_EnvVarToString(t *testing.T) {
	e := EnvironmentVariable{"DOJO_USER", "555", false, false}
	str := e.String()
	assert.Equal(t, "DOJO_USER=555", str)
}

func Test_multiLineVariablesToString(t *testing.T) {
	allVariables := []EnvironmentVariable{
		EnvironmentVariable{"DOJO_USER", "555", false, false},
		EnvironmentVariable{"DOJO_WORK_INNER", "/my/dir", false, false},
		EnvironmentVariable{"MULTI_LINE", `one
two`, true, false},
		EnvironmentVariable{"MYVAR", "999", false, false},
		EnvironmentVariable{"MULTI_LINE2", `one
two
three`, true, false},
	}
	genStr := multiLineVariablesToString(allVariables)
	assert.NotContains(t, genStr, "DOJO_USER")
	assert.NotContains(t, genStr, "MYVAR")
	assert.NotContains(t, genStr, "DOJO_WORK_INNER")
	assert.Contains(t, genStr, "export MULTI_LINE=$(echo b25lCnR3bw== | base64 -d)\n")
	assert.Contains(t, genStr, "export MULTI_LINE2=$(echo b25lCnR3bwp0aHJlZQ== | base64 -d)\n")
}

func Test_bashFunctionsVariablesToString(t *testing.T) {
	allVariables := []EnvironmentVariable{
				EnvironmentVariable{"DOJO_USER", "555", false, false},
				EnvironmentVariable{"MULTI_LINE", `one
		two`, true, false},
		EnvironmentVariable{"BASH_FUNC_my_bash_func%%", `() {
  echo "hello"
  echo "hi"
}`, true, true},
		EnvironmentVariable{"BASH_FUNC_my_bash_func2%%", `() {
  echo "abc"
}`, true, true},
		}
	genStr := bashFunctionsVariablesToString(allVariables)
	assert.NotContains(t, genStr, "DOJO_USER")
	assert.NotContains(t, genStr, "MULTI_LINE")
	assert.Contains(t, genStr, `my_bash_func() {
  echo "hello"
  echo "hi"
}
export -f my_bash_func
my_bash_func2() {
  echo "abc"
}
export -f my_bash_func2`)
}

//func Test_specialCharactersVariablesToString(t *testing.T) {
//	allVariables := []EnvironmentVariable{
//		EnvironmentVariable{"DOJO_USER", "555", false, false},
//		EnvironmentVariable{"MULTI_LINE", `one
//two`, true, false},
//		EnvironmentVariable{"MYVAR", "999", false, false},
//		EnvironmentVariable{"BASH_FUNC_bats_readlinkf%%", `() {  readlink -f "$1"
//second line
//}`, true, false},
//	}
//	genStr := multiLineVariablesToString(allVariables)
//	assert.NotContains(t, genStr, "DOJO_USER")
//	assert.NotContains(t, genStr, "MYVAR")
//	assert.Contains(t, genStr, "export MULTI_LINE=$(echo b25lCnR3bw== | base64 -d)\n")
//	assert.Contains(t, genStr, "export BASH_FUNC_bats_readlinkf%%=$(echo KCkgeyAgcmVhZGxpbmsgLWYgIiQxIgpzZWNvbmQgbGluZQp9 | base64 -d)\n")
//}


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
func (f *MockedEnvService) AddVariable(keyValue string){
	f.Variables = append(f.Variables, keyValue)
}