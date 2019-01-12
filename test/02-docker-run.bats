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
  assert_output --partial "alpine:3.8 \"whoami\""
  assert_output --partial "Exit status: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDockerContainerIsRemoved
}
@test "driver: docker, action: run, exit status: not 0" {
  cleanUpDockerContainer
  cleanUpEnvFiles
  run ${DOJO_PATH} --debug=true --test=true --image=alpine:3.8 notexistentcommand
  assert_output --partial "Dojo version"
  assert_output --partial "executable file not found"
  assert_output --partial "Exit status: 127"
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
  assert_output --partial "Exit status: 0"
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
  assert_output --partial "alpine:3.8 \"whoami\""
  assert_output --partial "Exit status: 0"
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
  assert_output --partial "Exit status: 0"
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
  assert_output --partial "Exit status: 0"
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