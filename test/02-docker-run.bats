load '/opt/bats-support/load.bash'
load '/opt/bats-assert/load.bash'
load 'common'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "driver: docker, action: run, exit status: 0" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "alpine:3.8 whoami"
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
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
  assert_equal "$status" 0
  assert_equal "${my_stdout}" "/root"
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
  assert_output --partial "Dojo version"
  assert_output --partial "executable file not found"
  assert_output --partial "Exit status from run command: 127"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  assert_equal "$status" 127
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command: unset" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 -i=false
  assert_output --partial "Dojo version"
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command: set after --" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 -i=false -- whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "alpine:3.8 whoami"
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command with quotes" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 sh -c "echo hello"
  assert_output --partial "Dojo version"
  assert_output --partial "alpine:3.8 sh -c \"echo hello\""
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, command: set after -- with quotes" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 -i=false --docker-options="--entrypoint=sh" -- -c "whoami"
  assert_output --partial "Dojo version"
  assert_output --partial "alpine:3.8 -c whoami"
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, environment preserved" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run bash -c "export ABC=custom_value ; ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 sh -c 'env | grep ABC'"
  assert_output --partial "Dojo version"
  assert_output --partial "custom_value"
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run; custom, relative work directory" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} -c test/test-files/Dojofile --debug=true --test=true --image=alpine:3.8 whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "alpine:3.8 whoami"
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, --rm=false" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  rm -f dojorc dojorc.txt
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 --rm=false whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "Exit status from run command: 0"
  assert_output --partial "Exit status from cleaning: 0"
  assert_output --partial "Exit status from signals: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  run cat dojorc
  assert_output --partial "DOJO_RUN_ID=testdojorunid"
  run cat dojorc.txt
  assert_output --partial "testdojorunid"
  assert_equal "$status" 0

  # test that environment file is NOT removed
  run bash -c "ls -la /tmp/ | grep 'test-dojo'"
  assert_equal "$status" 0

  # test that docker container is NOT removed
  run docker ps -a --filter "name=testdojorunid"
  assert_output --partial "testdojorunid"
  assert_equal "$status" 0
}
@test "clean" {
  cleanUpDockerContainer
  rm -f dojorc dojorc.txt
  rm -f /tmp/test-dojo*
}