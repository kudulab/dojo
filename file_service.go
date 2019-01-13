package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

type FileServiceInterface interface {
	GetFileUid(filePath string) uint32
	RemoveFile(filePath string, ignoreNoSuchFileError bool)
	ReadFile(filePath string) string
	ReadDockerComposeFile(filePath string) string
	WriteToFile(filePath string, contents string, logLevel string)
	GetCurrentDir() string
	RemoveGeneratedFile(removeContainers string, filePath string)
	RemoveGeneratedFileIgnoreError(removeContainers string, filePath string, ignoreNoSuchFileError bool)
	FileExists(filePath string) bool
}

type FileService struct {}

func (f FileService) GetFileUid(filePath string) uint32 {
	if filePath == "" {
		panic("filePath was empty")
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}
	mode := fileInfo.Sys().(*syscall.Stat_t)
	uid := mode.Uid
	return uid
}

func (f FileService) RemoveFile(filePath string, ignoreNoSuchFileError bool) {
	if filePath == "" {
		panic("filePath was empty")
	}
	err := os.Remove(filePath)
	if err != nil {
		if strings.Contains(err.Error(),"no such file or directory") && ignoreNoSuchFileError {
			Log("debug", fmt.Sprintf("File already removed: %s", filePath))
			return
		}
		panic(err)
	}
	Log("debug", fmt.Sprintf("Removed file: %s", filePath))
}
func (f FileService) ReadFile(filePath string) string {
	if filePath == "" {
		panic("filePath was empty")
	}
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(contents)
}
func (f FileService) FileExists(filePath string) bool {
	_, err := os.Lstat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(fmt.Sprintf("error when running os.Lstat(%q): %s", filePath, err))
	}
	return true
}

func (f FileService) ReadDockerComposeFile(filePath string) string {
	if filePath == "" {
		panic("filePath was empty")
	}
	return f.ReadFile(filePath)
}

func (f FileService) WriteToFile(filePath string, contents string, logLevel string) {
	if filePath == "" {
		panic("filePath was empty")
	}
	err := ioutil.WriteFile(filePath, []byte(contents), 0644)
	if err != nil {
		panic(err)
	}
	Log(logLevel, fmt.Sprintf("Written file %s, contents:\n %s", filePath, contents))
}

func (f FileService) GetCurrentDir() string {
	currentDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return currentDirectory
}
func (f FileService) RemoveGeneratedFile(removeContainers string, filePath string) {
	if filePath == "" {
		panic("filePath was empty")
	}
	if removeContainers != "false" {
		err := os.Remove(filePath)
		if err != nil {
			panic(err)
		}
		Log("debug", fmt.Sprintf("Removed generated file: %s", filePath))
		return
	} else {
		Log("debug", fmt.Sprintf("Not removed generated file: %s, because RemoveContainers is set", filePath))
		return
	}
}
func (f FileService) RemoveGeneratedFileIgnoreError(removeContainers string, filePath string, ignoreNoSuchFileError bool) {
	if filePath == "" {
		panic("filePath was empty")
	}
	if removeContainers != "false" {
		err := os.Remove(filePath)
		if err != nil {
			if strings.Contains(err.Error(),"no such file or directory") && ignoreNoSuchFileError {
				Log("debug", fmt.Sprintf("File already removed: %s", filePath))
				return
			}
			panic(err)
		}
		Log("debug", fmt.Sprintf("Removed file: %s", filePath))
		return
	} else {
		Log("debug", fmt.Sprintf("Not removed generated file: %s, because RemoveContainers is set", filePath))
		return
	}
}