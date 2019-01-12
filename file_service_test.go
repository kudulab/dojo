package main

import (
	"fmt"
)

type MockedFileService struct {}
func (f MockedFileService) GetFileUid(filePath string) uint32 {
	return 1000
}

func (f MockedFileService) RemoveFile(filePath string, ignoreNoSuchFileError bool) {
	Log("debug", fmt.Sprintf("Pretending to remove file: %s", filePath))
	return
}

func (f MockedFileService) ReadFile(filePath string) string {
	Log("debug", fmt.Sprintf("Pretending to read file %s", filePath))
	return ""
}

func (f MockedFileService) ReadDockerComposeFile(filePath string) string {
	Log("debug", fmt.Sprintf("Pretending to read file %s, returning a constant string", filePath))
	fileContents := `version: '2.2'
services:
  default:
    init: true
`
	return fileContents
}

func (f MockedFileService) WriteToFile(filePath string, contents string, logLevel string) {
	Log(logLevel, fmt.Sprintf("Pretending to write to file %s, contents:\n %s", filePath, contents))
	return
}

func (f MockedFileService) GetCurrentDir() string {
	return "/tmp/"
}
func (f MockedFileService) RemoveGeneratedFile(removeContainers string, filePath string) {
	if removeContainers != "false" {
		Log("debug", fmt.Sprintf("Pretending to remove generated file: %s", filePath))
	}
	return
}

