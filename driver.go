package main

import "fmt"

type DojoDriverInterface interface {
	HandleRun(mergedConfig Config, runID string, envService EnvServiceInterface) int
	CleanAfterRun(mergedConfig Config, runID string) int
	HandlePull(mergedConfig Config) int
	HandleSignal(mergedConfig Config, runID string) int
	HandleMultipleSignal(mergedConfig Config, runID string) int
}

func warnGeneral(fileService FileServiceInterface, config Config, envService EnvServiceInterface, logger *Logger) {
	if fileService.FileExists(config.WorkDirOuter) {
		if fileService.GetFileUid(config.WorkDirOuter) == 0 {
			logger.Log("warn", fmt.Sprintf(
				"WorkDirOuter: %s is owned by root, which is not recommended", config.WorkDirOuter))
		}
	} else {
		logger.Log("warn", fmt.Sprintf(
			"WorkDirOuter: %s does not exist", config.WorkDirOuter))
	}
	if !fileService.FileExists(config.IdentityDirOuter) {
		logger.Log("warn", fmt.Sprintf(
			"IdentityDirOuter: %s does not exist", config.IdentityDirOuter))
	}
	if envService.IsCurrentUserRoot() {
		logger.Log("warn", "Current user is root, which is not recommended")
	}
}