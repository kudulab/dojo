load 'common'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "driver: docker, action: run, exit status: 0" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 whoami
  [[ "$output" =~ 'Dojo version' ]]
  [[ "$output" =~ 'root' ]]
  [[ "$output" =~ 'alpine:3.8 whoami' ]]
  [[ "$output" =~ 'Exit status from run command: 0' ]]
  [[ "$output" =~ 'Exit status from cleaning: 0' ]]
  [[ "$output" =~ 'Exit status from signals: 0' ]]
  [[ ! "$output" =~ 'WARN' ]]
  [[ ! "$output" =~ 'warn' ]]
  [[ ! "$output" =~ 'ERROR' ]]
  [[ ! "$output" =~ 'error' ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}

@test "driver: docker, action: run, command output can be saved into a variable" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  rm -f stderr_file.txt
  # "What you cannot do is capture stdout in one variable, and stderr in another, using only FD redirections. You must use a temporary file (or a named pipe) to achieve that one."
  # http://mywiki.wooledge.org/BashFAQ/002
  my_stdout=$(${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 sh -c "printenv HOME" 2>stderr_file.txt)
  status=$?
  my_stderr=$(cat stderr_file.txt)
  [ "$status" -eq 0 ]
  [[ "${my_stdout}" = "/root" ]]
  echo "${my_stderr}" | grep "Exit status from run command: 0"
  echo "${my_stderr}" | grep "Exit status from cleaning: 0"
  echo "${my_stderr}" | grep "Exit status from signals: 0"
  echo "${my_stderr}" | grep "Dojo version"
  rm -f stderr_file.txt
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, exit status: not 0" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 notexistentcommand
  [[ "$output" =~ 'Dojo version' ]]
  [[ "$output" =~ 'executable file not found' ]]
  [[ "$output" =~ 'Exit status from run command: 127' ]]
  [[ ! "$output" =~ 'WARN' ]]
  [[ ! "$output" =~ 'warn' ]]
  [ "$status" -eq 127 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command: unset" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 -i=false
  [[ "$output" =~ 'Dojo version' ]]
  [[ "$output" =~ 'Exit status from run command: 0' ]]
  [[ "$output" =~ 'Exit status from cleaning: 0' ]]
  [[ "$output" =~ 'Exit status from signals: 0' ]]
  [[ ! "$output" =~ 'WARN' ]]
  [[ ! "$output" =~ 'warn' ]]
  [[ ! "$output" =~ 'ERROR' ]]
  [[ ! "$output" =~ 'error' ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command: set after --" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 -i=false -- whoami
  [[ "$output" =~ "Dojo version"  ]]
  [[ "$output" =~ "root"  ]]
  [[ "$output" =~ "alpine:3.8 whoami"  ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ "$output" =~ "Exit status from cleaning: 0" ]]
  [[ "$output" =~ "Exit status from signals: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command with quotes" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 sh -c "echo hello"
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "alpine:3.8 sh -c \"echo hello\"" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ "$output" =~ "Exit status from cleaning: 0" ]]
  [[ "$output" =~ "Exit status from signals: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command: set after -- with quotes" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 -i=false --docker-options="--entrypoint=sh" -- -c "whoami"
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "alpine:3.8 -c whoami" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ "$output" =~ "Exit status from cleaning: 0" ]]
  [[ "$output" =~ "Exit status from signals: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, environment preserved" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run bash -c "export ABC=custom_value ; ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 sh -c 'env | grep ABC'"
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "custom_value" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ "$output" =~ "Exit status from cleaning: 0" ]]
  [[ "$output" =~ "Exit status from signals: 0" ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run; custom, relative work directory" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} -c test/test-files/Dojofile --debug=true --test=true --image=alpine:3.8 whoami
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "root" ]]
  [[ "$output" =~ "alpine:3.8 whoami" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ "$output" =~ "Exit status from cleaning: 0" ]]
  [[ "$output" =~ "Exit status from signals: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run; custom, relative work directory - does not exist" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  rm -rf ./test/test-files/not-existent
  run ${DOJO_PATH} -c test/test-files/Dojofile.work_not_exists --debug=true --test=true --image=alpine:3.8 whoami
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "root" ]]
  [[ "$output" =~ "alpine:3.8 whoami" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ "$output" =~ "Exit status from cleaning: 0" ]]
  [[ "$output" =~ "Exit status from signals: 0" ]]
  [[ "$output" =~ "test/test-files/not-existent does not exist" ]]
  [[ "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
  rm -rf ./test/test-files/not-existent
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, --rm=false" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  rm -f dojorc dojorc.txt
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 --rm=false whoami
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "root" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ "$output" =~ "Exit status from cleaning: 0" ]]
  [[ "$output" =~ "Exit status from signals: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
  run cat dojorc
  [[ "$output" =~ "DOJO_RUN_ID=testdojorunid" ]]
  run cat dojorc.txt
  [[ "$output" =~ "testdojorunid" ]]
  [ "$status" -eq 0 ]

  # test that environment file is NOT removed
  run bash -c "ls -la /tmp/ | grep 'test-dojo'"
  [ "$status" -eq 0 ]

  # test that docker container is NOT removed
  run docker ps -a --filter "name=testdojorunid"
  [[ "$output" =~ "testdojorunid" ]]
  [ "$status" -eq 0 ]
}
@test "clean" {
  cleanUpDockerContainer
  rm -f dojorc dojorc.txt
  rm -f /tmp/test-dojo*
}
