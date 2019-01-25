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
