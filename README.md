# Dojo

Dojo helps you to execute builds and operations in [docker](https://docker.com) containers.

The Dojo project consists of:
 * `dojo` - a golang executable, which leverages `docker` and `docker-compose`
 * A specification and helper scripts for building [*dojo* docker images](#docker-images)

## Why?

Docker and containerization has revolutionized application lifecycle by creating a common standard for server applications. Every server software can be released as a docker image.
Dojo builds a similar standard for development environments, every environment should be versioned and released as a docker image.

Why Dojo exists and what problems it solves?
 * Dojo was created to **shorten setup time of development environment** to near-zero.
 * Dojo makes **executing builds in docker** much easier than bare `docker run` by taking care of many little details. (See [lower why](#Why-not-just-docker-run))
 * Dojo lets each project define its **development environment as code** and lock it in source control like a dependency. See [Dojofile](#dojofile)

Problems solved by Dojo and surrounding practices:
1. **Works on my machine** - with Dojo, each host (a developer's workstation or a CI agent) gets a **consistent, reproducible** environment delivered by a docker image.
1. **Configuration management** - with Dojo, any development environment is built from a [Dockerfile](https://docs.docker.com/engine/reference/builder/), which is versioned and released as a docker image.  Therefore the provisioning of the environment is stated in several commands, providing both documentation and a setup script in the single source of truth.
1. **Tools** on CI agents or developer's laptop over time may **mismatch with project's required setup** - With Dojo the **environment is pulled as docker image** just before the operation to be executed, therefore hosts don't need to be provisioned repeatedly to keep up with changing projects.

## Who is it for?

1. **as-code practitioners** - If you like *as-code philosophy* (infrastructure-as-code, pipelines-as-code) then you'll like Dojo, which provides **development and operations environment as code**.
1. Consultants, DevOps Engineers, **anyone jumping between projects and languages often**. With Dojo, getting the *right environment* for a project takes as long as the `docker pull`.
1. **Polyglot software houses**, especially where CI agents are shared by a variety of projects.
1. If you are already running builds with docker, then Dojo will make your scripts shorter and your life easier.

By adopting Dojo you would be shifting responsibility for the development environment closer to the developers.
A dojo docker image becomes a contract of what is a **correct environment** for a project **at a particular commit** (see [Dojofile](#dojofile)).

# Table of contents

1. [Quickstart](#quickstart)
1. [Installation](#installation)
1. [Dojofile](#dojofile)
1. [Docker images](#docker-images)
    * [Requirements](#image-requirements-and-best-practices)
    * [Building images](#image-scripts) with `Dockerfile`
        * Typical Dockerfile for [debian and ubuntu](#typical-debian-dockerfile)
        * Typical Dockerfile for [alpine](#typical-alpine-dockerfile)
1. [Secrets distribution](#secrets)
1. [Drivers](#drivers)
    * [docker](#docker-driver)
    * [docker-compose](#docker-compose-driver)
1. [Behavior reference](#behavior-reference)
    * [CLI arguments](#cli-arguments)
    * [Home and identity directory](#home-and-identity-directory)
    * [Handling signals](#handling-signals)
1. [FAQ](#faq)
1. [Comparison to other tools](#comparison-to-other-tools)
1. [Development](#Development)
1. [License](#License)

# Quickstart

1. [Install docker](https://docs.docker.com/install/), if you haven't already.
2. Install Dojo, it is a self-contained binary, so just place it somewhere on the `PATH`.

#### On Linux
```bash
DOJO_VERSION=0.6.0
wget -O dojo https://github.com/kudulab/dojo/releases/download/${DOJO_VERSION}/dojo_linux_amd64
sudo mv dojo /usr/local/bin
sudo chmod +x /usr/local/bin/dojo
```

#### On OSX
```bash
DOJO_VERSION=0.6.0
wget -O dojo https://github.com/kudulab/dojo/releases/download/${DOJO_VERSION}/dojo_darwin_amd64
mv dojo /usr/local/bin
chmod +x /usr/local/bin/dojo
```

Done. Now you have sufficient environment to build any project which leverages dojo images.

Let's build this [java project](https://github.com/tomzo/gocd-yaml-config-plugin) using dojo:

```bash
git clone https://github.com/tomzo/gocd-yaml-config-plugin.git
cd gocd-yaml-config-plugin
dojo "gradle test jar"
```

The output should start with a docker command that was executed:
```console
tomzo@073c1c477b1f:/tmp/gocd-yaml-config-plugin$ dojo "gradle test jar"
2019/04/28 14:01:40 [ 1]  INFO: (main.main) Dojo version 0.3.2
2019/04/28 14:01:40 [17]  INFO: (main.DockerDriver.HandleRun) docker command will be:
 docker run --rm -v /tmp/gocd-yaml-config-plugin:/dojo/work -v /home/tomzo:/dojo/identity:ro --env-file=/tmp/dojo-environment-dojo-gocd-yaml-config-plugin-2019-04-28_14-01-40-89074376 -v /tmp/.X11-unix:/tmp/.X11-unix -ti --name=dojo-gocd-yaml-config-plugin-2019-04-28_14-01-40-89074376 kudulab/openjdk-dojo:1.0.1 "gradle test jar"
Unable to find image 'kudulab/openjdk-dojo:1.0.1' locally
1.0.1: Pulling from kudulab/openjdk-dojo
```
Then you should see `docker pull` output and after creating the container, all tests being executed.
The build artifacts land in `build/libs/`, because that is how gradle behaves. These artifacts are available on our docker host.
Things to notice:
 * We have pulled [`kudulab/openjdk-dojo` docker image](https://github.com/kudulab/docker-openjdk-dojo) and created container from it.
 * Current directory `/tmp/gocd-yaml-config-plugin` was [mounted](https://docs.docker.com/storage/volumes/) to `/dojo/work`
 * Home directory on host `/home/tomzo` was [mounted](https://docs.docker.com/storage/volumes/) as readonly to `/dojo/identity`
 * After `gradle test jar` has finished, the docker container has exited and was removed. This is because we provided a command to `dojo`.

Every dojo image supports an interactive mode. In the above example, we can also run `dojo` at the root of the project to get dropped into an interactive shell.
Then we can work in the container for longer time, very much like in a [vagrant VM](#vs-vagrant).
Thanks to the mounted directory from the host, we can work on the project files using any other tools on our host, while container with java tools shares the same files.

This is a quickstart preview to give you a sense of how you would work with dojo. A more complete guide is being developed in this [readme](#table-of-contents) and an upcoming documentation site.

# Installation

### Prerequisites

You must be able to run a local docker daemon that can execute linux docker containers.
In practice this means Dojo works on **Linux or Mac**.
Dojo is continuously tested only on Linux.

### Dependencies
* Bash
* [Docker](https://docs.docker.com/)
* [Docker-Compose](https://github.com/docker/compose) >=1.7.1 (only if using Dojo driver: docker-compose)

### The binary
There is only 1 binary file to install:
```sh
version="0.6.0"
# on Linux:
wget -O /tmp/dojo https://github.com/kudulab/dojo/releases/download/${version}/dojo_linux_amd64
# or on Darwin:
# wget -O /tmp/dojo https://github.com/kudulab/dojo/releases/download/${version}/dojo_darwin_amd64
chmod +x /tmp/dojo
mv /tmp/dojo /usr/bin/dojo
```

Alternatively, you can build the binary file yourself, see [Development](#development).

# Docker images

A dojo docker image is responsible for setup of the development environment when container is created.

You can find complete examples of [open source "dojo images" on github](https://github.com/topics/dojo-image). Or you can build your own image starting from minimal example [snippets below](#image-scripts).

### Image requirements and best practices
We have several **rules and recommendations** on how to properly craft a dojo docker image. Dojo also provides [common setup scripts](#image-scripts) to achieve most of it.
1. Image must have `dojo` user and group. We recommend to set uid and gid to `1000`, although it is not necessary.
1. Image must have `bash` and `sudo` installed.
1. On container start, image should change uid and gid of dojo user to match with owner of the `/dojo/work` mount. Then, if any files were owned by dojo, their permissions should be updated.
1. The [docker image entrypoint](https://docs.docker.com/engine/reference/builder/#entrypoint):
    * must change current user to dojo user
    * should support running interactive and non-interactive shell
    * should source files in `/etc/dojo.d/variables/*`
    * should run files in `/etc/dojo.d/scripts/*`
    * should reap processes properly, by using an init system like [tini](https://github.com/krallin/tini) or [s6](https://github.com/just-containers/s6-overlay).
1. Image must ensure that current directory after starting container is `${dojo_work}` (by default `/dojo/work`).
1. Image shell prompt may include name of the dojo image being used. E.g. `dojo@297ea3cf20eb(openjdk-dojo):/dojo/work$`

We have also established several **best practices** for dojo image development:
 * Treat each dojo image as any software project with its own lifecycle. That implies: automate builds and releases, but most importantly **test the image before a release**.
 * Use semantic versioning for the docker image.
 * Leverage images available on dockerhub in the [`FROM` dockerfile instruction](https://docs.docker.com/engine/reference/builder/#from), then provision your custom tools and dojo-specific scripts. E.g. start a [java-dojo](https://github.com/kudulab/docker-openjdk-dojo/blob/master/image/Dockerfile#L1) from the official `openjdk` image.
 * If the development environment requires [secrets setup](#secrets), then container should assert validity of the secrets on start and exit with non-zero status if anything is missing.

## Image scripts

Dojo provides [several scripts](image_scripts/src) to be used inside dojo images to meet most of above requirements. Scripts can be installed in a `Dockerfile` with:

```dockerfile
ENV DOJO_VERSION=0.6.0
RUN git clone --depth 1 -b ${DOJO_VERSION} https://github.com/kudulab/dojo.git /tmp/dojo_git &&\
  /tmp/dojo_git/image_scripts/src/install.sh && \
  rm -r /tmp/dojo_git
```
Above script takes care of:
 * `dojo` user setup
 * handling owner of `/dojo/work`
 * provisioning of a correct [entrypoint](image_scripts/src/entrypoint.sh)

Image developer is still responsible for:
 * installing bash and sudo
 * ensuring that current directory after starting container is `${dojo_work}` (by default `/dojo/work`).

Below are valid, minimal examples of dojo dockerfiles:
 - [debian and ubuntu](#typical-debian-dockerfile)
 - [alpine](#typical-alpine-dockerfile)

### Typical debian dockerfile

For debian-family systems (including ubuntu) a typical dockerfile has following structure:

```dockerfile
FROM debian:9

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

# Install common Dojo scripts
ENV DOJO_VERSION=0.6.0
RUN apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
  sudo git ca-certificates && \
  git clone --depth 1 -b ${DOJO_VERSION} https://github.com/kudulab/dojo.git /tmp/dojo_git &&\
  /tmp/dojo_git/image_scripts/src/install.sh && \
  rm -r /tmp/dojo_git

#TODO: Add the tools your project needs

# Optional scripts to run on container start
#COPY etc_dojo.d/scripts/* /etc/dojo.d/scripts/
# Optional environment variables to source on container start
#COPY etc_dojo.d/variables/* /etc/dojo.d/variables/

COPY profile /home/dojo/.profile
COPY bashrc /home/dojo/.bashrc
RUN chown dojo:dojo /home/dojo/.profile /home/dojo/.bashrc

ENTRYPOINT ["/tini", "-g", "--", "/usr/bin/entrypoint.sh"]
CMD ["/bin/bash"]
```

The supporting `bashrc` can be used to setup a command prompt:
```bash
PS1='\[\033[01;32m\]\u@\h\[\033[00m\](minimal-debian-dojo):\[\033[01;34m\]\w\[\033[00m\]\$ '
```
And supporting `profile` can be used to ensure current directory is `${dojo_work}`:
```bash
# if running bash
if [ -n "$BASH_VERSION" ]; then
    # include .bashrc if it exists
    if [ -f "$HOME/.bashrc" ]; then
      . "$HOME/.bashrc"
    fi
fi
# this variable is set by common Dojo image scripts
cd "${dojo_work}"
```

### Typical alpine dockerfile

For alpine images a typical dockerfile has following structure:

```dockerfile
FROM alpine:3.9

# Install common Dojo scripts
ENV DOJO_VERSION=0.6.0
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
  apk add --no-cache tini bash shadow sudo git && \
  git clone --depth 1 -b ${DOJO_VERSION} https://github.com/kudulab/dojo.git /tmp/dojo_git &&\
  /tmp/dojo_git/image_scripts/src/install.sh && \
  rm -r /tmp/dojo_git

#TODO: Add the tools your project needs

# Optional scripts to run on container start
#COPY etc_dojo.d/scripts/* /etc/dojo.d/scripts/
# Optional environment variables to source on container start
#COPY etc_dojo.d/variables/* /etc/dojo.d/variables/

COPY profile /home/dojo/.profile
COPY bashrc /home/dojo/.bashrc
RUN chown dojo:dojo /home/dojo/.profile /home/dojo/.bashrc

ENTRYPOINT ["/sbin/tini", "-g", "--", "/usr/bin/entrypoint.sh"]
CMD ["/bin/bash"]
```

The supporting `bashrc` can be used to setup a command prompt:
```bash
PS1='\[\033[01;32m\]\u@\h\[\033[00m\](minimal-alpine-dojo):\[\033[01;34m\]\w\[\033[00m\]\$ '
```
And supporting `profile` can be used to ensure current directory is `${dojo_work}`:
```bash
# if running bash
if [ -n "$BASH_VERSION" ]; then
    # include .bashrc if it exists
    if [ -f "$HOME/.bashrc" ]; then
      . "$HOME/.bashrc"
    fi
fi
# this variable is set by common Dojo image scripts
cd "${dojo_work}"
```

## Secrets

Every organization manages secrets distribution differently. Dojo is agnostic towards the secret management. This section is an overview of several possibilities on how to deliver secrets into dojo containers.

#### Don't add secrets to the image

I hope this is obvious.

#### Use environment variables

By default every `dojo` run will pass environment variables from the host to the container. So if you already have infrastructure in place which delivers secrets as environment variables, these secrets will just work inside dojo containers.

#### Copy secrets from home directory

Dojo [mounts current user's home into `/dojo/identity`](#home-and-identity-directory), one of the reasons is to allow the container startup scripts to copy secrets from the user's home. For example, ssh keys could be setup by adding a following script in the dojo image:
`/etc/dojo.d/scripts/10-setup-ssh.sh`
```bash
#/bin/bash -e
if [ ! -d "${dojo_identity}/.ssh" ]; then
  echo "${dojo_identity}/.ssh does not exist. This image requires ssh keys for operation"
  exit 5
else
  cp -r "${dojo_identity}/.ssh/" "${dojo_home}/"
  find ${dojo_home}/.ssh -name '*id_rsa' -exec chmod -c 0600 {} \;
  find ${dojo_home}/.ssh -name '*id_rsa' -exec chown dojo:dojo {} \;
fi
# we need to ensure that ${dojo_home}/.ssh/config contains at least:
# StrictHostKeyChecking no
echo "StrictHostKeyChecking no
UserKnownHostsFile /dev/null
" > "${dojo_home}/.ssh/config"
```

#### Install secrets distribution tool in the image

All of your dojo images could include a tool such as [Vault](https://www.vaultproject.io/), which would be used in the [image startup scripts](#image-scripts) to provision secrets locally in the container. This however, still requires to use any of the other methods to deliver `VAULT_TOKEN` used to authorize with the vault server to obtain other secrets.

#### Mount secrets as a volume

We don't recommend this, but it is an option. If the secret is outside the home directory of the user running `dojo`, then it can be mounted into the container by using [additional docker options](#docker-options):
```toml
DOJO_DOCKER_OPTIONS="-v /path/to/local/secret:/path/in/docker"
```

# Dojofile

`Dojofile` is a file specifying the docker image which should be used for current project.
In simplest form `Dojofile` contains only `DOJO_DOCKER_IMAGE` with a docker image reference:
```toml
DOJO_DOCKER_IMAGE="image_name[:tag]"
```
For example, a java project might have following `Dojofile`:
```toml
DOJO_DOCKER_IMAGE="kudulab/openjdk-dojo:1.0.1"
```

We recommend to:
 * add `Dojofile` at the root of the project
 * add `Dojofile` to the source control
 * use unambiguous docker tags, such as `kudulab/openjdk-dojo:1.0.1` rather than `kudulab/openjdk-dojo:latest`. This guarantees that current commit will be always built in the same image, which helps with reproducible builds.

### Dojofile options

`Dojofile` has several settings to control `dojo` behavior.

##### Dojo driver

```toml
DOJO_DRIVER="docker"
```
Defines which driver to use. Possible values are [docker](#docker-driver) or [docker-compose](#docker-compose-driver). Default is `docker`.

*equivalent CLI option is: `--driver`*

##### Image

```toml
DOJO_DOCKER_IMAGE="kudulab/dotnet-dojo:3.1.0"
```
Defines which image to use. The value must be a valid docker image reference, same as you would specify in `docker pull`. There is no default, you must specify an image in Dojofile or in CLI arguments.

*equivalent CLI option is: `--image`*

##### Docker options

```toml
DOJO_DOCKER_OPTIONS="-p 9090:80 --privileged"
```
Defines additional arguments for the `docker run` command. Default is empty.

*equivalent CLI option is: `--docker-options`*

##### Outer working directory

```toml
DOJO_WORK_OUTER="/tmp/my-project"
```
Defines directory on the docker host to mount into the container at `/dojo/work` (by default).
The default is to mount current working directory.

`Dojo` setups following mount `DOJO_WORK_OUTER:DOJO_WORK_INNER`.

*equivalent CLI option is: `--work-dir-outer`*

We do not recommend changing this.

##### Inner working directory

```toml
DOJO_WORK_INNER="/dojo/work/my-project"
```
Defines directory inside the docker container, which is mounted from the host. By default `/dojo/work`.

`Dojo` setups following mount `DOJO_WORK_OUTER:DOJO_WORK_INNER`.

*equivalent CLI option is: `--work-dir-inner`*

We do not recommend changing this.

##### Identity directory

```toml
DOJO_IDENTITY_OUTER="/home/joe"
```
Defines directory on the docker host to mount into the container at `/dojo/identity`.
The default is to mount current user's home. See more about [identity directory](#home-and-identity-directory).

`Dojo` setups following mount `DOJO_IDENTITY_OUTER:/dojo/identity`.

*equivalent CLI option is: `--identity-dir-outer`*

##### Blacklist variables

```toml
DOJO_BLACKLIST_VARIABLES="HOME,USER,MY_PREFIX_*"
```
By default dojo will pass environment variables from host to the container.
`DOJO_BLACKLIST_VARIABLES` can be used to control which variables should not be mapped.
It is a list of environment variables, split by commas, which should not be transfered from host to the docker container.
A blacklisted variable name can end with an asterisk, e.g. `BASH*`, which means that all the variables with BASH prefix (like `BASH_123`, `BASHBLA`) will be blacklisted.

The default value is:
```
"BASH*,HOME,USERNAME,USER,LOGNAME,PATH,TERM,SHELL,MAIL,SUDO_*,WINDOWID,SSH_*,SESSION_*,GEM_HOME,GEM_PATH,GEM_ROOT,HOSTNAME,HOSTTYPE,IFS,PPID,PWD,OLDPWD,LC*,TMPDIR"
```

*equivalent CLI option is: `--blacklist`*

##### Log level

```toml
DOJO_LOG_LEVEL="info"
```

*In CLI use `--debug=true` or `--debug=false`*


##### Docker-compose file

```toml
DOJO_DOCKER_COMPOSE_FILE="my-docker-compose.yaml"
```
Used only with [docker-compose driver](#docker-compose-driver), points to the docker-compose file which should be used for creating containers when running `docker-compose run`.
Default is `docker-compose.yml`.

*equivalent CLI option is: `-docker-compose-file`*

##### Docker-compose options

```toml
DOJO_DOCKER_COMPOSE_OPTIONS="--service-ports"
```
Defines additional arguments for the `docker-compose run` command. Used only with [docker-compose driver](#docker-compose-driver). Default is empty.

*no equivalent in CLI*

##### Docker-compose exit behavior

```toml
DOJO_EXIT_BEHAVIOR="abort"
```
Only applicable for [docker-compose driver](#docker-compose-driver).
Defines how to react when a non-default container exits. Possible values are:
 * `ignore` - the docker-compose run continues until default container exits.
 * `abort` (default) - the docker-compose run will be interrupted by dojo once any container stops.
 * `restart` - dojo will restart any non-default container which has stopped.

*equivalent CLI option is: `-exit-behavior`*

# Drivers

Dojo can run commands with [docker](#docker-driver) or [docker-compose](#docker-compose-driver), which is controlled by [`DOJO_DRIVER` option in dojofile](#dojo-driver).

## docker driver

`docker` driver is the default. It is based on `docker run` command and allows to create just one container for the development environment. It is useful for running builds, unit tests, simple CLI tools.

## docker-compose driver

`docker-compose` driver is based on `docker-compose run`. Several containers can be created to setup the development environment.
`Dojo` treats first container as the *default* (primary) container. It is the only container which has to be dojo-compliant. That means, you can use any other images in the remaining services.

Dojo with `docker-compose` driver is a very powerful tool. Some examples when it can be useful:
 * developing a java application which connects to a database. You can create `java-dojo` and linked `postgresql` container with a database server.
 * developing microservices, where default container is a dojo container where you develop an application, while several linked applications are containers with dependant microservices.
 * integration and performance testing. Especially with the `abort` [exit behavior](#docker-compose-exit-behavior), failure of any container will stop and fail the test.
 * Real-world example usage is [LiGet](https://github.com/ai-traders/LiGet), where we test nuget clients against a nuget server running in docker.

### docker-compose file

In order to use `docker-compose` driver, a Dojofile would include:
```toml
DOJO_DRIVER="docker-compose"
DOJO_DOCKER_IMAGE="kudulab/openjdk-dojo:1.0.1"
DOJO_DOCKER_COMPOSE_FILE="docker-compose.yml"
```

Next to the `Dojofile`, you must create a docker-compose.yml according the [official reference](https://docs.docker.com/compose/compose-file/).
Example docker-compose file:
```yaml
version: '2.2'
services:
  default:
    links:
      - db:db
  db:
    init: true
    image: postgres:11.2-alpine
    environment:
      - POSTGRES_PASSWORD=my_pw
```

The docker-compose file must meet following several requirements to work with Dojo.
 * version of docker-compose file must be >=2.2
 * there must be a default service declared - it will be the container running a dojo docker image.
 * do not set `image` option in the default service. Because Dojo sets it based on `DOJO_DOCKER_IMAGE` from `Dojofile` or using CLI option.

You can try creating above 2 files in any directory and run `dojo`. The output should look like this:

```console
tomzo@073c1c477b1f:/tmp/compose-example$ dojo
2019/04/28 18:38:17 [ 1]  INFO: (main.main) Dojo version 0.3.2
2019/04/28 18:38:18 [20]  INFO: (main.DockerComposeDriver.HandleRun) docker-compose run command will be:
 docker-compose -f docker-compose.yml -f docker-compose.yml.dojo -p dojo-compose-example-2019-04-28_18-38-17-33362956 run --rm default
Creating network "dojo-compose-example-2019-04-28_18-38-17-33362956_default" with the default driver
Pulling db (postgres:11.2-alpine)...
11.2-alpine: Pulling from library/postgres
bdf0201b3a05: Already exists
365f27dc05d7: Pull complete
bf541d40dfbc: Pull complete
823ce70c3252: Pull complete
a92a31ecd32a: Pull complete
83cc8c6d8282: Pull complete
7995b9edc9bf: Pull complete
7616119153d9: Pull complete
b3f69561e369: Pull complete
Creating dojo-compose-example-2019-04-28_18-38-17-33362956_db_1 ... done
28-04-2019 18:38:27 Dojo entrypoint info: Sourcing: /etc/dojo.d/variables/50-variables.sh
28-04-2019 18:38:27 Dojo entrypoint info: Sourcing: /etc/dojo.d/variables/61-java-variables.sh
28-04-2019 18:38:27 Dojo entrypoint info: Sourcing: /etc/dojo.d/scripts/20-setup-identity.sh
28-04-2019 18:38:28 Dojo entrypoint info: Sourcing: /etc/dojo.d/scripts/50-fix-uid-gid.sh
+ usermod -u 1000 dojo
usermod: no changes
+ groupmod -g 1000 dojo
+ chown 1000:1000 -R /home/dojo
28-04-2019 18:38:28 Dojo entrypoint info: dojo init finished (interactive shell)
dojo@297ea3cf20eb(openjdk-dojo):/dojo/work$
```

# Behavior reference

### CLI arguments
```
$ dojo --help
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
  -preserve-env-to-all string

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


### Home and identity directory

By default Dojo mounts current user's home directory into `/dojo/identity`. The mount is read-only.
There are several reasons for such design:
 * `/dojo/identity` is read-only so hacking with the project inside container cannot break your home directory.
 * If user's home was mounted into `/home/dojo`, then many of the user's application settings could unintentionally break configuration of the dojo user in container. Instead every dojo image can selectively use `/dojo/identity` to setup `/home/dojo` so that container becomes operational.

### Handling signals
Dojo reacts on signals: SIGINT (Ctrl+C) and SIGTERM. Dojo recognizes if 1 or 2
signals were sent. This means: you can press e.g. Ctrl+C two times to speed up containers stopping.

| event | driver | Dojo behavior | exit status |
| --- | --- | --- | --- |
| one signal | docker | docker stop that 1 container | 130 for SIGINT, 2 for SIGTERM |
| one signal | docker-compose | docker stop default container, then docker-compose stop to gracefully stop all the other containers | 130 for SIGINT, 2 for SIGTERM |
| 2 signals | docker | docker kill that 1 container | 3 |
| 2 signals | docker-compose | docker kill default container, then docker-compose kill to immediately stop all the other containers | 3 |
| >2 signals | all | ignored | 3 |


In order to not couple Dojo behavior with Docker and Docker-Compose behavior, signals from terminal are not simply preserved onto `docker run` or `docker-compose run` process (the process is started with a separate process group ID).


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

# FAQ

### Will Dojo work with any docker image?

It will run, but this is not recommended and won't be very convenient because:
 * Most images by default use `root` user inside the image, therefore any files created in the container will be owned by root.
 * `Dojo` mounts current directory into `/dojo/work`, a correct dojo image would by default enter `/dojo/work`. A random image usually enters `/`.
 * A correct dojo image would setup current user inside docker container to match with UID and GID of the owner of `/dojo/work` mount. A random image usually runs as root, even if not, then still chances of UID/GID match with `/dojo/work` are low.

### Why not just docker run?

You may be using `docker run` or `docker-compose run` already for your builds. For example: `docker run -ti --volume $PWD:/build openjdk:8u212 gradle test`.
There are several issues which are not solved out of the box by these tools, in fact all of them are reasons why dojo was created:
 * Not running builds as root
 * Interactive vs non-interactive shell problems. E.g. you could not run `docker run -ti` on a CI agent, but you can on interactive terminal.
 * *Selectively* passing environment variables from host to the container.
 * Handling mismatch between UID/GID of the user on the host and UID/GID inside the container. Basically keeping artifacs and project files owned by the same user during the entire time.
 * In docker-compose dealing with containers crashed during the run
 * Capturing and handling signals.
 * Returning exit code from inside container to the host.
 * Cleanup of the containers and networks.
 * [Dojofile](#dojofile) enforces versioning of the reference to development environment.

### Multiple dojofiles

A single project can use multiple dojo images, therefore multiple `Dojofiles` should be checked-in the repository. For example, front-end and backend can live in same repository, written in 2 different languages.

We recommend to name your dojofiles with a suffix suggesting the type of environment. E.g. `Dojofile-nodejs` and `Dojofile-python`.
To point which `Dojofile` should be used, pass `-c` option to `dojo`:
```bash
dojo -c Dojofile-nodejs npm install
```

## Comparison to other tools

### vs. Batect

[Batect](https://github.com/charleskorn/batect) is the closest tool to Dojo and its created for the same reasons.
The primary differences between dojo and Batect:
 - Batect requires java to run, Dojo is a self-contained binary
 - Batect orchestrates multiple containers by itself, dojo uses docker-compose
 - Batect is configured with YAML where you define lifecycle tasks of the project. Dojo is agnostic towards the lifecycle, you can use any existing tool (Make, gradle, bash scripts) to define tasks/targets of the project.
 - Batect encourages usage of existing official images, therefore YAML config of each project has to define options such as volume mounts, users, etc. Dojo encourages having custom images, so that Dojofile config is minimal within a project.
 - Dojo does not support Windows.

### vs. Vagrant

[Vagrant](https://www.vagrantup.com/) has the same mission as Dojo - to make portable, easy to setup, development environments.
However vagrant is a tool from the *virtualization era* and is built around virtual machines. The typical workflow in vagrant is
```
vagrant up
vagrant ssh
vagrant down
```
Major difference is that:
 * vagrant assumes a long running development environment, rather then spinning up new host for running each command. Dojo allows to have both approaches.
 * `vagrant up` can take time (VM booting time)
 * vagrant (in most cases) requires a periodic synchronization of your current working directory to the VM and back. In dojo, the filesystem is shared between the container and the host.

Still vagrant is a great tool and irreplaceable if you must use a VM.

`Dojofile` is similar to `Vagrantfile` in its function. But with a restriction that Dojo forces you to first build and release the image, while `Vagrantfile` has some tools to provision the VM when running `vagrant up` which adds up to the startup time.

### vs. dependency managers

A typical dependency manager allows you to commit references to all immediate and transitive dependencies in source control, files such as `Gemfile.lock`, `paket.lock`, `yarn.lock` etc. It is a good practice to do so, because project defines fully, in code, what libraries it depends on.
If we extend this concept, then we soon realize that the project did not define how the host, where it should be built, should look like. Probably, the most common situation is that there is a README file listing several tools to install. It does not provide a good developer experience.
In any sensible organization, configuration management is used to provision the CI-agents and developer's workstation building the project. It is one level better than the README, but still the definition of the environment exists outside the project and there is no reference to it.
The `Dojofile` is the *lock* file which allows to store this reference in source control. When you run `dojo <command>`, then the environment is fetched, just like a dependency manager would fetch all dependencies of the project.

## Development
Run these commands in `dojo` (use previous version of dojo to build a next one):
```
./tasks deps
./tasks build
./tasks unit
```

Run integration tests, (this uses [inception-dojo](https://github.com/kudulab/docker-inception-dojo) docker image):
```
./tasks e2e alpine
./tasks e2e ubuntu18
```

## License

Copyright 2019 Ewa Czechowska, Tomasz Sętkowski

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
