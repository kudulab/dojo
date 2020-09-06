### 0.10.0 (2020-Sep-06)

* Added support for [homebrew on Linux](https://github.com/kudulab/dojo/pull/20). Thanks to [Justin Garrison](https://github.com/rothgar)

### 0.9.0 (2020-Aug-13)

* support exported bash functions [#17](https://github.com/kudulab/dojo/issues/17).
   Earlier, Dojo resulted in an error when any bash function was exported. Now,
   it will succeed. However, in order to preserve all the exported bash functions, you
   need to run:
   ```
   source /etc/dojo.d/variables/01-bash-functions.sh
   ```

### 0.8.0 (2020-Jan-01)

* Docker-composer driver: enable printing logs of non default docker containers either to console or to file.
 Previously, only printing on console (stderr) was supported. This can be set by `--print-logs-target`
 commandline option and by `DOJO_DOCKER_COMPOSE_PRINT_LOGS_TARGET` Dojofile option.
 Possible values: console, file. [#12](https://github.com/kudulab/dojo/issues/12)
* Fix automated github releases to add release notes and to not mark a release as pre-release

### 0.7.0 (2020-Jan-01)

Docker-compose driver: print logs of non default docker containers. By default this will be done if any
 of the containers (default or not) failed. This can be set by `--print-logs` commandline option
 and by `DOJO_DOCKER_COMPOSE_PRINT_LOGS` Dojofile option.
 Possible values: always, failure, never. [#12](https://github.com/kudulab/dojo/issues/12)

### 0.6.3 (2019-Sep-19)

Fix homebrew formula to handle upgrades

### 0.6.2 (2019-Sep-19)

Added homebrew tap automation

### 0.6.1 (2019-Sep-16)

Blacklist TMPDIR to improve OSX experience

### 0.6.0 (2019-Sep-13)

Added OSX support

### 0.5.0 (2019-May-03)

* support multi-line environment variables [#4](https://github.com/ai-traders/dojo/issues/4).
 From now on, we will pass into --env-file only these environment variables, which value is a single line.
 The variables which values are multi-line, are now saved to another file on docker host and mounted onto
 docker container(s) as /etc/dojo.d/variables/00-multiline-vars.sh. The entrypoint is expected to source
 all files in /etc/dojo.d/variables.
 In order to not deal with escaping quotes or special characters, the multi-line variables' values are
 serialized with base64. Example of a multi-line variable serialized in 00-multiline-vars.sh:
 ```
 export ABC=$(echo MTExCjIyMiAzMzM= | base64 -d)
 ```

### 0.4.3 (2019-May-01)

* fix: added script to image_scripts which always setups `/run/user/<ID>`;
 add tests
* image scripts: if olduid == newuid, do not run usermod or groupmod

### 0.4.2 (2019-Apr-30)

* use the same environment variables when running a `docker run` or `docker-compose run` command and
when generating the envFile

### 0.4.1 (2019-Apr-29)

* do not run chown dojo home dir when uid/gid matches

### 0.4.0 (2019-Apr-27)

* added e2e tests on alpine and ubuntu18, executed in `inception-dojo` image
* ported tests to pytest, dropped bats which does not work on alpine
* export `DOJO_WORK_*` variables for all started processes \#17391
* added script to image_scripts which always setups `/run/user/<ID>`

### 0.3.2 (2019-Apr-22)

* cross compile on Linux and Darwin
* add Darwin support for verification if shell is interactive, thanks to [#2](https://github.com/ai-traders/dojo/pull/2), [@Eiffel-Alpine](https://github.com/Eiffel-Alpine)
* from now on release two binaries: one for Linux amd64 and one for Darwin amd64

### 0.3.1 (2019-Feb-04)

* fix: while saving the environment variable: DISPLAY a new line was missing

### 0.3.0 (2019-Feb-04)

Change default behavior that preserved current environment (by the means of environment file) only to the default
container (driver: docker-compose). Now, by default we preserve to all the containers. Still the old behavior can
be set by:
```
dojo --preserve-env-to-all=false
```


### 0.2.1 (2019-Feb-01)

* fix: resolve relative paths in config object
* fix: allow docker and docker-compose run if WorkDirOuter or IdentityDirOuter does not exist
* make github release happen on CI

### 0.2.0 (2019-Jan-25)

* feature: support configurable exit behavior for docker-compose driver if a non-default container stops #17197
* fix: better react on signals, do not depend on how docker and docker-compose react

### 0.1.0 (2019-Jan-16)

Initial release. Features #17138:
   1. Support drivers: docker, docker-compose.
   1. Read configuration from CLI and configuration file. Default configuration file is `./Dojofile`. Dojofile is not a bash script, it will be not sourced.
   1. Support running without any configuration file.
   1. Support actions: `dojo -a run` (default) and `dojo -a pull`.
   1. Support forcing not interactive docker run and docker-compose run.
   1. Support not removing docker containers (only for driver: docker).
   1. Support blacklisting env variables. If any env variable is blacklisted, it will be kept with "DOJO_" prefix. A good default of blacklisted variables is set.
   1. Support graphical mode (opinionated right now). If environment variable `DISPLAY` is set, mount as docker volume: `/tmp/.X11-unix:/tmp/.X11-unix`.
   1. If Dojo is run as root, show a warning.
   1. If DOJO_WORK_OUTER is owned by root, show a warning.
   1. Support ctrl+c (SIGINT) and SIGTERM.
   1. Provide example Dojo docker image entrypoint, test it on Alpine and Ubuntu.
   1. Support easy to configure docker-compose files. Do not require end user to add lines like `${DOJO_IDENTITY}:/dojo/identity:ro`,
      Dojo handles that by generating another docker-compose file (`docker-compose.yml.dojo`).
      Thus, we need docker-compose >=1.7.1, because it added [multiple compose files support](https://docs.docker.com/compose/extends/#multiple-compose-files)
