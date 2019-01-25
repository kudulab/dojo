# Dojo

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

Alternatively, you can build the binary file yourself, see [Development](#Development).

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

### How to test that Dojo reacts on signals
Dojo reacts on signals: SIGINT (Ctrl+C) and SIGTERM. Dojo recognizes if 1 or 2 signals were sent. This means: you can
 press e.g. Ctrl+C two times to speed up containers stopping.

#### Container's PID 1 process **not** preserving signals
Run this
```
dojo --debug=true --test=true --image=alpine:3.8 -i=false sh -c "echo 'will sleep' && sleep 1d"
```
and after it prints "will sleep", press Ctrl+C.
Dojo will wait for the container to be removed by docker, but it will be not removed (because sh does not
preserves signals). Then, Dojo will stop the container. Exit status will be: 130.

You can press Ctrl+C again, this will invoke `docker kill` and exit immediately. Exit status will be: 3.

#### Container's PID 1 process preserving signals
Run this
```
dojo --docker-options="--init" --debug=true --test=true --image=alpine:3.8 -i=false sh -c "echo 'will sleep' && sleep 1d"
```
and after it prints "will sleep", press Ctrl+C.
Dojo will wait for the container to be removed by docker, it will removed. Exit status will be: 130.


#### Driver: docker-compose
Run this:
```
dojo --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true -i=false --image=alpine:3.8 sh -c "echo 'will sleep' && sleep 1d"
```
and after it prints "will sleep", press Ctrl+C. Docker-compose reacts fast on the 2nd Ctrl+C.

In order to make docker-compose default container preserve signals, let your docker-compose file contain:
```
version: '2.2'
services:
  default:
    init: true
```
Minimal versions that support this feature: version `2.2` of docker-compose file and version `1.7.1` of docker-compose application.

#### Implementation
| event | driver | what to do |
| --- | --- | --- |
| one signal | docker | docker stop that 1 container |
| one signal | docker-compose | docker stop default container, then docker-compose stop to gracefully stop all the other containers (there may be order by which the rest of containers should be stopped) |
| 2 signals | docker | docker kill that 1 container |
| 2 signals | docker-compose | docker kill default container, then docker-compose kill to immediately stop all the other containers (there may be order by which the rest of containers should be stopped) |
| >2 signals | all | ignored |

Signals are not preserved: the processes `docker run` and `docker-compose run` are started with a separate process group ID.
`docker-compose stop` and `docker-compose kill` do not stop the default container, created with `docker-compose run default CMD`,
thus we first stop the default container.
After goroutines that handle signals are finished, cleanup is performed (remove docker containers, network, environment file
 and `docker-compose.yml.dojo` file).