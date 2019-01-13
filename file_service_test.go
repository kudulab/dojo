package main

import (
	"fmt"
)

type MockedFileService struct {
	FilesWrittenTo map[string]string
	FilesRemovals  []string
}
func NewMockedFileService() *MockedFileService {
	mm := make(map[string]string, 0)
	rm := make([]string, 0)
	return &MockedFileService{
		FilesWrittenTo: mm,
		FilesRemovals:  rm,
	}
}

func (f *MockedFileService) GetFileUid(filePath string) uint32 {
	return 1000
}

func (f *MockedFileService) RemoveFile(filePath string, ignoreNoSuchFileError bool) {
	f.FilesRemovals = append(f.FilesRemovals, filePath)
	Log("debug", fmt.Sprintf("Pretending to remove file: %s", filePath))
	return
}

func (f *MockedFileService) ReadFile(filePath string) string {
	Log("debug", fmt.Sprintf("Pretending to read file %s", filePath))
	return ""
}

func (f *MockedFileService) FileExists(filePath string) bool {
	Log("debug", fmt.Sprintf("Pretending that file exists %s", filePath))
	return true
}

func (f *MockedFileService) ReadDockerComposeFile(filePath string) string {
	Log("debug", fmt.Sprintf("Pretending to read file %s, returning a constant string", filePath))
	fileContents := `version: '2.2'
services:
  default:
    init: true
`
	return fileContents
}

func (f *MockedFileService) WriteToFile(filePath string, contents string, logLevel string) {
	f.FilesWrittenTo[filePath] = contents
	Log(logLevel, fmt.Sprintf("Pretending to write to file %s, contents:\n %s", filePath, contents))
	return
}

func (f *MockedFileService) GetCurrentDir() string {
	return "/tmp/"
}
func (f *MockedFileService) RemoveGeneratedFile(removeContainers string, filePath string) {
	if removeContainers != "false" {
		f.FilesRemovals = append(f.FilesRemovals, filePath)
		Log("debug", fmt.Sprintf("Pretending to remove generated file: %s", filePath))
	}
	return
}
func (f *MockedFileService) RemoveGeneratedFileIgnoreError(removeContainers string, filePath string, ignoreError bool) {
	if removeContainers != "false" {
		f.FilesRemovals = append(f.FilesRemovals, filePath)
		Log("debug", fmt.Sprintf("Pretending to remove generated file: %s", filePath))
	}
	return
}

