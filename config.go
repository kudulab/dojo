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
	RunCommand         string

	DockerImage string
	DockerOptions string
	DockerComposeFile string
	DockerComposeOptions string
	Test string
}

func (c Config) String() string {
	str := ""
	str += fmt.Sprintf("{ Action: %s }", c.Action)
	str += fmt.Sprintf("{ ConfigFile: %s }", c.ConfigFile)
	str += fmt.Sprintf("{ Driver: %s }", c.Driver)
	str += fmt.Sprintf("{ Debug: %s }", c.Debug)
	str += fmt.Sprintf("{ Dryrun: %s }", c.Dryrun)
	str += fmt.Sprintf("{ Interactive: %s }", c.Interactive)
	str += fmt.Sprintf("{ RemoveContainers: %s }", c.RemoveContainers)
	str += fmt.Sprintf("{ WorkDirInner: %s }", c.WorkDirInner)
	str += fmt.Sprintf("{ WorkDirOuter: %s }", c.WorkDirOuter)
	str += fmt.Sprintf("{ IdentityDirOuter: %s }", c.IdentityDirOuter)
	str += fmt.Sprintf("{ BlacklistVariables: %s }", c.BlacklistVariables)
	str += fmt.Sprintf("{ RunCommand: %s }", c.RunCommand)
	str += fmt.Sprintf("{ DockerImage: %s }", c.DockerImage)
	str += fmt.Sprintf("{ DockerOptions: %s }", c.DockerOptions)
	str += fmt.Sprintf("{ DockerComposeFile: %s }", c.DockerComposeFile)
	str += fmt.Sprintf("{ DockerComposeOptions: %s }", c.DockerComposeOptions)
	str += fmt.Sprintf("{ Test: %s }", c.Test)
	return str
}

func getCLIConfig() Config {
	// let's use use a custom flagSet, so that we don't mutate global state
	flagSet := flag.NewFlagSet("flagSet", flag.PanicOnError)

	var help bool
	const usageHelp      = "Print help and exit 0"
	flagSet.BoolVar(&help, "help", false, usageHelp)
	flagSet.BoolVar(&help, "h", false, usageHelp+" (shorthand)")

	var version bool
	const usageVersion      = "Print version and exit 0"
	flagSet.BoolVar(&version, "version", false, usageVersion)
	flagSet.BoolVar(&version, "v", false, usageVersion+" (shorthand)")

	var action string
	const usageAction      = "Action: run, pull. Default: run"
	flagSet.StringVar(&action, "action", "", usageAction)
	flagSet.StringVar(&action, "a", "", usageAction+" (shorthand)")

	var config string
	const usageConfig         = "Config file. Default: ./Dojofile"
	flagSet.StringVar(&config, "config", "", usageConfig)
	flagSet.StringVar(&config, "c", "", usageConfig+" (shorthand)")

	var driver string
	const usageDriver         = "Driver: docker or docker-compose (dc for short). Default: docker"
	flagSet.StringVar(&driver, "driver", "", usageDriver)
	flagSet.StringVar(&driver, "d", "", usageDriver+" (shorthand)")

	var image string
	const usageImage         = "Docker image name and tag, e.g. alpine:3.8"
	flagSet.StringVar(&image, "image", "", usageImage)

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

	var dockerComposeFile string
	const usageDCFile         = "Docker-compose file. Default: ./docker-compose.yml. Only for driver: docker-compose"
	flagSet.StringVar(&dockerComposeFile, "docker-compose-file", "", usageDCFile)
	flagSet.StringVar(&dockerComposeFile, "dcf", "", usageDCFile+" (shorthand)")

	var test string
	const usageTest         = "Set this to true when integration testing. This turns writing env files to a test directory"
	flagSet.StringVar(&test, "test", "", usageTest)

	flagSet.Parse(os.Args[1:])
	runCommandArr := flagSet.Args()
	runCommand := smartJoinCommandArgs(runCommandArr)

	flagSet.Usage = func () {
		fmt.Fprint(os.Stderr, "Usage of dojo <flags> [--] <CMD>:\n")
		flagSet.PrintDefaults()
	}
	if help {
		flagSet.Usage()
		os.Exit(0)
	}
	if version {
		// version is returned as the first log message
		os.Exit(0)
	}

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
		RunCommand:         runCommand,
		DockerImage:        image,
		DockerComposeFile:  dockerComposeFile,
		Test: 				test,
	}
}

// While parsing CLI arguments, after all the flags are handled, we want to treat the rest of the arguments
// as 1 element, as docker or docker-compose run command.
// We cannot just use strings.Join(runCommandArr, " ") because this would result in missing quotes.
func smartJoinCommandArgs(commandArgs []string) string {
	quotedArgs := make([]string,0)
	for _,v := range commandArgs {
		//  It is safe to assume that an argument that contains white space(s) must have been (and should be) quoted
		//  http://stackoverflow.com/a/1669493/4457564
		//  Otherwise, this input command: -c "echo aaa"
		//  would result in an output command: -c echo aaa.
		updatedStr := v
		if strings.Contains(v, " ") {
			// If quotes were used 2 times, we have to escape the inner quotes. E.g. input command was:
			// "/bin/bash -c \"echo aaa\" && echo bbb".
			// Notice that, the shell strips outer quotes of an argument, so that argument (here: v) will
			// never have outer quotes. Thus, by checking if v contains quotes, we check for the quotes
			// which are not at the first or last string char (we check for the inner quotes).
			if strings.Contains(v, "\"") {
				updatedStr = strings.Replace(updatedStr, "\"", "\\\"", -1)
			}
			// and now let's quote it
			updatedStr = fmt.Sprintf("\"%s\"", updatedStr)
		}
		quotedArgs = append(quotedArgs, updatedStr)
	}
	return strings.Join(quotedArgs,  " ")
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
	config.RunCommand = configMap["runCommand"]
	config.DockerImage = configMap["dockerImage"]
	config.DockerOptions = configMap["dockerOptions"]
	config.DockerComposeFile = configMap["dockerComposeFile"]
	config.DockerComposeOptions = configMap["dockerComposeOptions"]
	config.Test = configMap["test"]
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
	configMap["runCommand"] = config.RunCommand
	configMap["dockerImage"] = config.DockerImage
	configMap["dockerOptions"] = config.DockerOptions
	configMap["dockerComposeFile"] = config.DockerComposeFile
	configMap["dockerComposeOptions"] = config.DockerComposeOptions
	configMap["test"] = config.Test
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
				case "DOJO_DOCKER_COMPOSE_OPTIONS":
					config.DockerComposeOptions = value
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
		DockerComposeFile:  "docker-compose.yml",
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

func verifyConfig(config *Config) error {
	if config.Driver == "dc" {
		config.Driver = "docker-compose"
	}
	if config.Action != "run" && config.Action != "pull" && config.Action != "help" && config.Action != "version" {
		return fmt.Errorf("Invalid configuration, unsupported Action: %s. Supported: run, pull", config.Action)
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
	if config.DockerComposeOptions != "" && config.Driver == "docker" {
		return fmt.Errorf("DockerComposeOptions option is unsupported for driver: docker")
	}
	if config.DockerOptions != "" && config.Driver == "docker-compose" {
		return fmt.Errorf("DockerOptions option is unsupported for driver: docker-compose")
	}
	if config.RemoveContainers == "false" && config.Driver == "docker-compose" {
		Log("warn", "RemoveContainers=false is unsupported for driver: docker-compose")
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
	if config.Driver == "docker-compose" {
		if config.DockerComposeFile == "" {
			return fmt.Errorf("Invalid configuration, DockerComposeFile is unset")
		}
		dcFile := config.DockerComposeFile
		if _, err := os.Stat(dcFile); err != nil {
			return fmt.Errorf("docker-compose config file: %s does not exist", dcFile)
		}
	}
	return nil
}