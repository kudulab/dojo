# Dojo

Runs an isolated, reproducible, well-defined environment for operations.

## Usage

### Examples
Run interactively (no command set), using configuration file.
```
$ cat Dojofile
DOJO_DOCKER_IMAGE="alpine:3.8"
```
```
$ dojo
2019/01/25 13:48:01 [ 1]  INFO: (main.main) Dojo version 0.1.0
2019/01/25 13:48:01 [17]  INFO: (main.DockerDriver.HandleRun) docker command will be:
 docker run --rm -v /home/ewa/code/dojo:/dojo/work -v /home/ewa:/dojo/identity:ro \
   --env-file=/tmp/dojo-environment-dojo-dojo-2019-01-25_13-48-01-17043942 \
   -ti --name=dojo-dojo-2019-01-25_13-48-01-17043942 "alpine:3.8"
/ # whoami
root
/ # exit
$ echo $?
0
```

Run one command, using configuration file:
```
$ cat Dojofile
DOJO_DOCKER_IMAGE="alpine:3.8"
```
```
$ dojo whoami
2019/01/25 13:49:08 [ 1]  INFO: (main.main) Dojo version 0.1.0
2019/01/25 13:49:08 [17]  INFO: (main.DockerDriver.HandleRun) docker command will be:
 docker run --rm -v /home/ewa/code/dojo:/dojo/work -v /home/ewa:/dojo/identity:ro \
   --env-file=/tmp/dojo-environment-dojo-dojo-2019-01-25_13-49-08-38384463 \
   -ti --name=dojo-dojo-2019-01-25_13-49-08-38384463 "alpine:3.8" whoami
root
$ echo $?
0
```

Run one command, without configuration file:
```
$ dojo --image=alpine:3.8 whoami
2019/01/25 13:50:37 [ 1]  INFO: (main.main) Dojo version 0.1.0
2019/01/25 13:50:37 [ 7]  INFO: (main.DockerDriver.HandleRun) docker command will be:
 docker run --rm -v /home/ewa/code/dojo:/dojo/work -v /home/ewa:/dojo/identity:ro \
   --env-file=/tmp/dojo-environment-dojo-dojo-2019-01-25_13-50-37-28841830 \
   -ti --name=dojo-dojo-2019-01-25_13-50-37-28841830 alpine:3.8 whoami
root
```

Run one command, save output to a variable:
```
$ my_stdout=$(dojo whoami)
2019/01/25 13:51:39 [ 1]  INFO: (main.main) Dojo version 0.1.0
2019/01/25 13:51:39 [17]  INFO: (main.DockerDriver.HandleRun) docker command will be:
 docker run --rm -v /home/ewa/code/dojo:/dojo/work -v /home/ewa:/dojo/identity:ro \
  --env-file=/tmp/dojo-environment-dojo-dojo-2019-01-25_13-51-39-19563225 \
  --name=dojo-dojo-2019-01-25_13-51-39-19563225 "alpine:3.8" whoami
$ echo $my_stdout
root
```

Run one command, using configuration file, exit status is non 0.
```
$ cat Dojofile
DOJO_DOCKER_IMAGE="alpine:3.8"
```
```
$ dojo not-existent-cmd
2019/01/25 14:00:58 [ 1]  INFO: (main.main) Dojo version 0.1.0
2019/01/25 14:00:58 [17]  INFO: (main.DockerDriver.HandleRun) docker command will be:
 docker run --rm -v /home/ewa/code/dojo:/dojo/work -v /home/ewa:/dojo/identity:ro \
   --env-file=/tmp/dojo-environment-dojo-dojo-2019-01-25_14-00-58-90652377 \
   -ti --name=dojo-dojo-2019-01-25_14-00-58-90652377 "alpine:3.8" not-existent-cmd
docker: Error response from daemon: OCI runtime create failed: container_linux.go:348: starting container process caused "exec: \"not-existent-cmd\": executable file not found in $PATH": unknown.
$ echo $?
127
```

Run one command, without configuration file, driver is docker-compose:
```
$ dojo --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --image=alpine:3.8 whoami
2019/01/25 13:53:03 [ 1]  INFO: (main.main) Dojo version 0.1.0
2019/01/25 13:53:03 [17]  INFO: (main.FileService.WriteToFile) Written file ./test/test-files/itest-dc.yaml.dojo, contents:
 version: '2.2'
services:
  default:
    image: alpine:3.8
    volumes:
      - /home/ewa:/dojo/identity:ro
      - /home/ewa/code/dojo:/dojo/work
      - /tmp/.X11-unix:/tmp/.X11-unix
    env_file:
      - /tmp/dojo-environment-dojo-dojo-2019-01-25_13-53-03-10131355
2019/01/25 13:53:03 [17]  INFO: (main.DockerComposeDriver.HandleRun) docker-compose run command will be:
 docker-compose -f ./test/test-files/itest-dc.yaml -f ./test/test-files/itest-dc.yaml.dojo -p dojo-dojo-2019-01-25_13-53-03-10131355 run --rm default whoami
Creating network "dojodojo2019012513530310131355_default" with the default driver
Creating dojodojo2019012513530310131355_abc_1 ...
Creating dojodojo2019012513530310131355_abc_1 ... done
root
2019/01/25 13:53:05 [17]  INFO: (main.DockerComposeDriver.stop) Stopping containers with command:
docker-compose -f ./test/test-files/itest-dc.yaml -f ./test/test-files/itest-dc.yaml.dojo -p dojo-dojo-2019-01-25_13-53-03-10131355 stop
Stopping dojodojo2019012513530310131355_abc_1 ... done
2019/01/25 13:53:06 [ 1]  INFO: (main.DockerComposeDriver.CleanAfterRun) Removing containers with command:
docker-compose -f ./test/test-files/itest-dc.yaml -f ./test/test-files/itest-dc.yaml.dojo -p dojo-dojo-2019-01-25_13-53-03-10131355 down
Removing dojodojo2019012513530310131355_abc_1 ... done
Removing network dojodojo2019012513530310131355_default
```

Just pull the images without creating docker containers:
```
$ dojo --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --action=pull
2019/01/25 14:20:13 [ 1]  INFO: (main.main) Dojo version 0.1.0
2019/01/25 14:20:13 [ 1]  INFO: (main.DockerComposeDriver.HandlePull) docker-compose pull command will be:
 docker-compose -f ./test/test-files/itest-dc.yaml -f ./test/test-files/itest-dc.yaml.dojo -p dojo pull
3.8: Pulling from library/alpine
```

### What is it?
Dojo is a CLI program. It spawns docker containers, invoking either `docker run`
 or `docker-compose run` command. You are responsible for providing
  a **specialized Dojo docker image**, but Dojo will work with any docker image.

Why not just use the official Docker images? Read: [Dojo Docker image](README-dojo-docker-image.md).

Dojo is responsible for constructing a proper `docker run` or `docker-compose run`
 command where:
   * it bind-mounts (rw) current directory (WorkDirOuter) on host onto `/dojo/work`
    directory (WorkDirInner) in docker container
   * it bind-mounts (ro) your HOME directory (IdentityDirOuter) on host onto
   `/dojo/identity` directory in docker container
   * it preserves environment variables, while blacklisting some of them, e.g.:
    HOME, USERNAME, USER, any variables starting with BASH prefix, etc. All the
    environment variables that are going to be preserved into a container are saved
    into a file: `/tmp/dojo-environment-<run_ID>`. Then, this file is passed to
    `docker run` command as `--env-file` value.
   * For docker-compose driver, second docker-compose file is generated. Its name
   ends with `.dojo` and it is used to set all the above pointed informations.
   * if you have `DISPLAY` environment variable set, then an opinionated method
   is used to ensure the container will work with graphical applications.
   (The method is: to mount `/tmp/.X11-unix:/tmp/.X11-unix` and to set this in the
   docker container: `DISPLAY=unix:0.0`).

After the command is finished, Dojo runs cleanup (docker containers are stopped and removed, temporary files are removed).

In order to understand the Dojo bind-mounts, read: [Dojo Docker image](README-dojo-docker-image.md).

### Configuration
You can keep configuration in a file, conventionally named: `Dojofile` or use CLI flags.
```
$ ./bin/dojo --help
Usage of dojo <flags> [--] <CMD>:
  -a string
    	Action: run, pull. Default: run (shorthand)
  -action string
    	Action: run, pull. Default: run
  -blacklist string
    	List of variables, split by commas, to be blacklisted in a docker container
  -c string
    	Config file. Default: ./Dojofile (shorthand)
  -config string
    	Config file. Default: ./Dojofile
  -d string
    	Driver: docker or docker-compose (dc for short). Default: docker (shorthand)
  -dcf string
    	Docker-compose file. Default: ./docker-compose.yml. Only for driver: docker-compose (shorthand)
  -debug string
    	Set log level to debug (verbose). Default: false
  -docker-compose-file string
    	Docker-compose file. Default: ./docker-compose.yml. Only for driver: docker-compose
  -docker-options string
    	Options to the docker run command. E.g. "--init"
  -driver string
    	Driver: docker or docker-compose (dc for short). Default: docker
  -exit-behavior string
    	How to react when a container (not the default one) exits. Possible values: ignore, abort (default), restart. Only for driver: docker-compose
  -h	Print help and exit 0 (shorthand)
  -help
    	Print help and exit 0
  -i string
    	Set to false if you want to force not interactive docker run
  -identity-dir-outer string
    	Directory on host, to be mounted into a docker container to /dojo/identity. Default: $HOME
  -image string
    	Docker image name and tag, e.g. alpine:3.8
  -interactive string
    	Set to false if you want to force not interactive docker run
  -remove-containers string
    	Set to true if you want to not remove docker containers. Default: true
  -rm string
    	Set to true if you want to not remove docker containers. Default: true
  -test string
    	Set this to true when integration testing. This turns writing env files to a test directory
  -v	Print version and exit 0 (shorthand)
  -version
    	Print version and exit 0
  -w string
    	Directory in a docker container, to which we bind mount from host. Default: /dojo/work (shorthand)
  -work-dir-inner string
    	Directory in a docker container, to which we bind mount from host. Default: /dojo/work
  -work-dir-outer string
    	Directory on host, to be mounted into a docker container. Default: current directory
```

All the configuration file options with example values:
```
# same as CLI: driver
DOJO_DRIVER="docker"
# same as CLI: image
DOJO_DOCKER_IMAGE="alpine:3.8"
# same as CLI: docker-options
DOJO_DOCKER_OPTIONS="-e ABC=123 --privileged"
# same as CLI: docker-compose-file
DOJO_DOCKER_COMPOSE_FILE="my-docker-compose.yaml"
# no counterpart in CLI. Options for docker-compose run command.
DOJO_DOCKER_COMPOSE_OPTIONS="--service-ports"
# same as CLI: work-dir-outer
DOJO_WORK_OUTER="/tmp/my-project"
# same as CLI: work-dir-inner
DOJO_WORK_INNER="/dojo/work/my-project"
# same as CLI: identity-dir-outer
DOJO_IDENTITY_OUTER="/tmp/my-identity"
# same as CLI: exit-behavior
DOJO_EXIT_BEHAVIOR="abort"
# same as CLI: blacklist
DOJO_BLACKLIST_VARIABLES="HOME,USER,MY_PREFIX_*"
# you can set in CLI: --debug=true or --debug=false
DOJO_LOG_LEVEL="info"
```

#### docker-compose file
If you choose driver: docker-compose, you have to provide a docker-compose file.
Requirements:
   * version of docker-compose file must be >=2.2
   * there must be a default service declared
   * do not set `image` option, it will be overridden anyways by the image you
   configured in configuration file or by CLI flag.

Dojo will run a command in the default container.

Example docker-compose file:
```
version: '2.2'
services:
  default:
    init: true
    links:
    - abc:abc
    volumes:
    - /tmp/additional-volume:/tmp/additional-volume
    environment:
      ABC_DEF: "1234"
  abc:
    init: true
    image: alpine:3.8
    entrypoint: ["/bin/sh", "-c"]
    command: ["while true; do sleep 1d; done;"]
```
By default, ExitBehavior is set to abort, which means that if any of the non-default
 containers is stopped (in the example above: abc container), Dojo will stop
 all the docker-compose containers.


#### Reacting on signals
Dojo reacts on signals: SIGINT (Ctrl+C) and SIGTERM. Dojo recognizes if 1 or 2
 signals were sent. This means: you can
press e.g. Ctrl+C two times to speed up containers stopping.

| event | driver | what to do | exit status |
| --- | --- | --- | --- |
| one signal | docker | docker stop that 1 container | 130 for SIGINT, 2 for SIGTERM |
| one signal | docker-compose | docker stop default container, then docker-compose stop to gracefully stop all the other containers (there may be order by which the rest of containers should be stopped) | 130 for SIGINT, 2 for SIGTERM |
| 2 signals | docker | docker kill that 1 container | 3 |
| 2 signals | docker-compose | docker kill default container, then docker-compose kill to immediately stop all the other containers (there may be order by which the rest of containers should be stopped) | 3 |
| >2 signals | all | ignored | 3 |


In order to not couple Dojo behavior with Docker and Docker-Compose behavior, signals from terminal are not simply preserved onto `docker run` or `docker-compose run`
 process (the process is started with a separate process group ID).


`docker-compose stop` and `docker-compose kill` do not stop the default container, created with `docker-compose run default CMD`,
thus, firstly, we stop the default container.


After goroutines that handle signals are finished, cleanup is performed (remove docker containers, network, environment file
and `docker-compose.yml.dojo` file).

To test reacting on signals, run the following commands. After "will sleep"
 is printed, press Ctrl+C. You may also test pressing Ctrl+C more times.

1. driver: docker, container's PID 1 process **not** preserving signals:
```
dojo --image=alpine:3.8 -i=false sh -c "echo 'will sleep' && sleep 1d"
```
2. driver: docker, container's PID 1 process preserving signals:
```
dojo --docker-options="--init" --image=alpine:3.8 -i=false sh -c "echo 'will sleep' && sleep 1d"
```
3. driver: docker-compose:
```
dojo --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml -i=false --image=alpine:3.8 sh -c "echo 'will sleep' && sleep 1d"
```

## Install
### Dependencies
* [Docker](https://docs.docker.com/)
* [Docker-Compose](https://github.com/docker/compose) (only if using Dojo driver: docker-compose) >=1.7.1
* Bash shell


### The binary
There is only 1 binary file to install:
```
version="0.2.0"
wget -O=/tmp/dojo https://github.com/ai-traders/dojo/releases/download/${version}/dojo
chmod +x /tmp/dojo
mv /tmp/dojo /usr/bin/dojo
```

Alternatively, you can build the binary file yourself, see [Development](#development).

## Aims
Dojo is intended to be used with **short-running commands**, e.g. to compile code, to create backup dump, to use a 3rd party software to do some operation like e.g. use [Terraform](https://www.terraform.io/) to create a VM or run some graphical IDE for local development. It is rather not expected to use Dojo with long-running commands, e.g. to run a server 24/7.

Dojo was invented in order to resolve the problems of:
   * **fast changing work environment requirements** (you had to reinstall/reconfigure over and over again)
   * using **different environments at the same moment** (e.g using Java 7 for some projects and Java 8 for other)
   * **work environments becoming bloated** (in one machine you have to have installed and configured: Mono, Java, Python, etc.)
   * **not reproducible work environments** - **works on my machine** (e.g. you installed Mono just fine on your work computer, but then installing it on your laptop demanded some additional steps or resulted in having installed different versions). This is especially important if you use  **Continuous Integration**, the environment should be the same on workstation and CI agents.
   * **mutable work environments** - e.g. your code compiled a month ago, but you installed sth new and it influenced the environment
   * **experimenting with new technologies without breaking your workstation**

## Limitations
Dojo was tested only on Linux, with local Docker server.

## Development
Run those command in `ide`:
```
./tasks deps
./tasks build
./tasks unit
```
run integration tests in environment with Bats installed:
```
./tasks e2e
```
