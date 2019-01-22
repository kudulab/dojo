package main

type DojoDriverInterface interface {
	HandleRun(mergedConfig Config, runID string, envService EnvServiceInterface) int
	CleanAfterRun(mergedConfig Config, runID string) int
	HandlePull(mergedConfig Config) int
	HandleSignal(mergedConfig Config, runID string) int
	HandleMultipleSignal(mergedConfig Config, runID string) int
}
