package main

type DojoDriverInterface interface {
	HandleRun(mergedConfig Config, runID string, envService EnvServiceInterface) int
	HandlePull(mergedConfig Config) int
	HandleSignal(mergedConfig Config, runID string) int
}
