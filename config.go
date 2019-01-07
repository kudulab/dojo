package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
)

type Config struct {
	Action             string
	ConfigFile         string
	Driver             string
	Debug              string
	Dryrun             string
	Interactive        string
	RemoveContainers   string
	WorkDirInner       string
	WorkDirOuter       string
	IdentityDirOuter   string
	BlacklistVariables string
	DockerRunCommand   string

	DockerImage string
	DockerOptions string
	DockerComposeFile string
}

func getCLIConfig() Config {
	// let's use use a custom flagSet, so that we don't mutate global state
	flagSet := flag.NewFlagSet("flagSet", flag.PanicOnError)

	var action string
	const usageAction      = "Action: run, pull, version, help. Default: run"
	flagSet.StringVar(&action, "action", "", usageAction)
	flagSet.StringVar(&action, "a", "", usageAction+" (shorthand)")

	var config string
	const usageConfig         = "Config file. Default: ./Dojofile"
	flagSet.StringVar(&config, "config", "", usageConfig)
	flagSet.StringVar(&config, "c", "", usageConfig+" (shorthand)")

	var driver string
	const usageDriver         = "Driver: docker or docker-compose. Default: docker"
	flagSet.StringVar(&driver, "driver", "", usageDriver)
	flagSet.StringVar(&driver, "d", "", usageDriver+" (shorthand)")

	// this is not bool, because we need to know if it was set or not
	var debug string
	const usageDebug         = "Set log level to debug (verbose). Default: false"
	flagSet.StringVar(&debug, "debug", "", usageDebug)

	// this is not bool, because we need to know if it was set or not
	var dryrun string
	const usageDryrun         = "Do not pull docker image, do not run docker. Default: false"
	flagSet.StringVar(&dryrun, "dryrun", "", usageDryrun)

	// this is not bool, because we need to know if it was set or not
	var interactive string
	const usageInteractive         = "Set to false if you want to force not interactive docker run"
	flagSet.StringVar(&interactive, "interactive", "", usageInteractive)
	flagSet.StringVar(&interactive, "i", "", usageInteractive)

	// this is not bool, because we need to know if it was set or not
	var removeContainers string
	const usageRm         = "Set to true if you want to not remove docker containers. Default: true"
	flagSet.StringVar(&removeContainers, "remove-containers", "", usageRm)
	flagSet.StringVar(&removeContainers, "rm", "", usageRm)

	var workDirInner string
	const usageWorkDirInner         = "Directory in a docker container, to which we bind mount from host. Default: /dojo/work"
	flagSet.StringVar(&workDirInner, "work-dir-inner", "", usageWorkDirInner)
	flagSet.StringVar(&workDirInner, "w", "", usageWorkDirInner+" (shorthand)")

	var workDirOuter string
	const usageWworkDirOuter        = "Directory on host, to be mounted into a docker container. Default: current directory"
	flagSet.StringVar(&workDirOuter, "work-dir-outer", "", usageWworkDirOuter)

	var identityDirOuter string
	const usageIdentityDirOuter = "Directory on host, to be mounted into a docker contaienr to /dojo/identity. Default: $HOME"
	flagSet.StringVar(&identityDirOuter, "identity-dir-outer", "", usageIdentityDirOuter)

	var blacklistVariables string
	const usageBlackilstVariables    = "List of variables, split by commas, to be blacklisted in a docker container"
	flagSet.StringVar(&blacklistVariables, "blacklist", "", usageBlackilstVariables)

	flagSet.Parse(os.Args[1:])
	dockerRunCommand := flagSet.Args()

	return Config{
		Action:             action,
		ConfigFile:         config,
		Driver:             driver,
		Debug:              debug,
		Dryrun:             dryrun,
		Interactive:        interactive,
		RemoveContainers:   removeContainers,
		WorkDirInner:       workDirInner,
		WorkDirOuter:       workDirOuter,
		IdentityDirOuter:   identityDirOuter,
		BlacklistVariables: blacklistVariables,
		DockerRunCommand:   strings.Join(dockerRunCommand, " "),
	}
}
func MapToConfig(configMap map[string]string) Config {
	config := Config{}
	config.Action = configMap["action"]
	config.ConfigFile = configMap["config"]
	config.Driver = configMap["driver"]
	config.Debug = configMap["debug"]
	config.Dryrun = configMap["dryrun"]
	config.Interactive = configMap["interactive"]
	config.RemoveContainers = configMap["removeContainers"]
	config.WorkDirInner = configMap["workDirInner"]
	config.WorkDirOuter = configMap["workDirOuter"]
	config.IdentityDirOuter = configMap["identityDirOuter"]
	config.BlacklistVariables = configMap["blacklistVariables"]
	config.DockerRunCommand = configMap["dockerRunCommand"]
	config.DockerImage = configMap["dockerImage"]
	config.DockerOptions = configMap["dockerOptions"]
	config.DockerComposeFile = configMap["dockerComposeFile"]
	return config
}
func ConfigToMap(config Config) map[string]string {
	configMap := make(map[string]string,0)
	configMap["action"] = config.Action
	configMap["config"] = config.ConfigFile
	configMap["driver"] = config.Driver
	configMap["debug"] = config.Debug
	configMap["dryrun"] = config.Dryrun
	configMap["interactive"] = config.Interactive
	configMap["removeContainers"] = config.RemoveContainers
	configMap["workDirInner"] = config.WorkDirInner
	configMap["workDirOuter"] = config.WorkDirOuter
	configMap["identityDirOuter"] = config.IdentityDirOuter
	configMap["blacklistVariables"] = config.BlacklistVariables
	configMap["dockerRunCommand"] = config.DockerRunCommand
	configMap["dockerImage"] = config.DockerImage
	configMap["dockerOptions"] = config.DockerOptions
	configMap["dockerComposeFile"] = config.DockerComposeFile
	return configMap
}

// getFileConfig never returns error. If config file does not exist,
// it returns Config object with default values.
func getFileConfig(filePath string) Config {
	config := Config{}
	if _, err := os.Stat(filePath); err == nil {
		contents, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Print(err)
		}
		lines := strings.Split(string(contents), "\n")

		for _, line := range lines {
			if !strings.HasPrefix(line,"#") && line != "" {
				// the file line is not a comment

				// there may be many "=" signs in this line, let's just consider the 1st one
				kv := strings.SplitN(line, "=", 2)
				key := kv[0]
				value := kv[1]

				switch key {
				case "DOJO_DRIVER":
					config.Driver = value
				case "DOJO_DOCKER_IMAGE":
					config.DockerImage = value
				case "DOJO_DOCKER_OPTIONS":
					config.DockerOptions = value
				case "DOJO_DOCKER_COMPOSE_FILE":
					config.DockerComposeFile = value
				case "DOJO_WORK_OUTER":
					config.WorkDirOuter = value
				case "DOJO_WORK_INNER":
					config.WorkDirInner = value
				case "DOJO_IDENTITY_OUTER":
					config.IdentityDirOuter = value
				case "DOJO_BLACKLIST_VARIABLES":
					config.BlacklistVariables = value
				case "DOJO_LOG_LEVEL":
					if value == "debug" || value == "DEBUG" {
						config.Debug = "true"
					} else {
						config.Debug = "false"
					}
				}
			}
		}
	} else {
		Log("debug", fmt.Sprintf("Config file does not exist: %s", filePath))
	}
	return config
}

func getDefaultConfig(configFile string) Config {
	currentDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	currentUser, err:= user.Current()
	if err != nil {
		panic(err)
	}
	defaultConfig := Config{
		Action:             "run",
		ConfigFile:         configFile,
		Driver:             "docker",
		Debug:              "false",
		Dryrun:             "false",
		RemoveContainers:   "true",
		WorkDirOuter:       currentDirectory,
		WorkDirInner:       "/dojo/work",
		IdentityDirOuter:   currentUser.HomeDir,
		BlacklistVariables: "BASH*,HOME,USERNAME,USER,LOGNAME,PATH,TERM,SHELL,MAIL,SUDO_*,WINDOWID,SSH_*,SESSION_*,GEM_HOME,GEM_PATH,GEM_ROOT,HOSTNAME,HOSTTYPE,IFS,PPID,PWD,OLDPWD,LC*",
	}
	return defaultConfig
}

func getMergedConfig(moreImportantConfig Config, lessImportantConfig Config, leastImportantConfig Config) Config {
	leastImportantConfigMap := ConfigToMap(leastImportantConfig)
	moreImportantConfigMap := ConfigToMap(moreImportantConfig)
	lessImportantConfigMap := ConfigToMap(lessImportantConfig)

	mergedConfigMap := make(map[string]string,0)
	for k,v := range moreImportantConfigMap {
		if v != "" {
			mergedConfigMap[k] = v
		} else if lessImportantConfigMap[k] != "" {
			mergedConfigMap[k] = lessImportantConfigMap[k]
		} else {
			mergedConfigMap[k] = leastImportantConfigMap[k]
		}
	}

	config := MapToConfig(mergedConfigMap)
	return config
}

func verifyConfig(config Config) error {
	if config.Action != "run" && config.Action != "pull" && config.Action != "help" && config.Action != "version" {
		return fmt.Errorf("Invalid configuration, unsupported Action: %s. Supported: run, pull, help, version", config.Action)
	}
	if config.Driver != "docker" && config.Driver != "docker-compose" {
		return fmt.Errorf("Invalid configuration, unsupported Driver: %s. Supported: docker, docker-compose", config.Driver)
	}
	if config.Debug != "true" && config.Debug != "false" {
		return fmt.Errorf("Invalid configuration, unsupported Debug: %s. Supported: true, false", config.Debug)
	}
	if config.Dryrun != "true" && config.Dryrun != "false" {
		return fmt.Errorf("Invalid configuration, unsupported Dryrun: %s. Supported: true, false", config.Dryrun)
	}
	if config.RemoveContainers != "true" && config.RemoveContainers != "false" {
		return fmt.Errorf("Invalid configuration, unsupported RemoveContainers: %s. Supported: true, false", config.RemoveContainers)
	}
	if config.Interactive != "true" && config.Interactive != "false" && config.Interactive != "" {
		return fmt.Errorf("Invalid configuration, unsupported Interactive: %s. Supported: true, false, empty string", config.Interactive)
	}
	if config.DockerImage == "" {
		return fmt.Errorf("Invalid configuration, DockerImage is unset")
	}
	return nil
}