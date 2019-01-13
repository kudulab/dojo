# Dojo

## Development
Run those command in `ide`:
```
./tasks deps
./tasks build
./tasks unit
```
run integration tests in environment with Bats installed:
```
./tasks itest
```

### How to test that Dojo reacts on signals
Dojo gets notified of signals: SIGINT (Ctrl+C) and SIGTERM. It respects multiple notifications (2). This means: you can
 press e.g. Ctrl+C two times.

#### Container's PID 1 process **not** preserving signals
Run this
```
dojo --debug=true --test=true --image=alpine:3.8 -i=false sh -c "echo 'will sleep' && sleep 1d"
```
and after it prints "will sleep", press Ctrl+C.
Dojo will wait for the container to be removed by docker, but it will be not removed (because sh does not
preserves signals). Then, Dojo will stop the container. Exit status will be: 130.

You can press Ctrl+C again, this will invoke `docker kill` and exit immediately.

#### Container's PID 1 process preserving signals
Run this
```
dojo --docker-options="--init" --debug=true --test=true --image=alpine:3.8 -i=false sh -c "echo 'will sleep' && sleep 1d"
```
and after it prints "will sleep", press Ctrl+C.
Dojo will wait for the container to be removed by docker, it will removed. Exit status will be: 130.