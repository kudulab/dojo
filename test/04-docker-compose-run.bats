load '/opt/bats-support/load.bash'
load '/opt/bats-assert/load.bash'
load 'common'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

function cleanUpDCDojoFile() {
  rm -f test/test-files/itest-dc.yaml.dojo
}

function testDCDojoFileIsRemoved(){
  run test -f test/test-files/itest-dc.yaml.dojo
  assert_equal "$status" 1
}

function cleanUpDCContainers() {
  docker stop testdojorunid_default_run_1 >/dev/null 2>&1 || true
  docker stop testdojorunid_abc_1 >/dev/null 2>&1 || true
  docker rm testdojorunid_default_run_1 >/dev/null 2>&1 || true
  docker rm testdojorunid_abc_1 >/dev/null 2>&1 || true
}

function testDCContainersAreRemoved() {
  run docker ps -a --filter "name=testdojorunid"
  refute_output --partial "testdojorunid"
  assert_equal "$status" 0
}

function cleanUpDCNetwork() {
  docker network rm testdojorunid_default || true
}

function testDCNetworkIsRemoved() {
  run docker network ls --filter "name=testdojorunid"
  refute_output --partial "testdojorunid"
  assert_equal "$status" 0
}

@test "driver: docker-compose, action: run, exit status: 0" {
  cleanUpDCContainers
  cleanUpDCNetwork
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "whoami"
  assert_output --partial "Exit status from run command: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
  testDCContainersAreRemoved
  testDCNetworkIsRemoved
}

@test "driver: docker-compose, action: run, command output can be saved into a variable" {
  cleanUpDCContainers
  cleanUpDCNetwork
  cleanUpEnvFiles
  cleanUpDCDojoFile
  rm -f stderr_file.txt
  # "What you cannot do is capture stdout in one variable, and stderr in another, using only FD redirections. You must use a temporary file (or a named pipe) to achieve that one."
  # http://mywiki.wooledge.org/BashFAQ/002
  my_stdout=$(${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 sh -c "printenv HOME" 2>stderr_file.txt)
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
  testDCDojoFileIsRemoved
  testDCContainersAreRemoved
  testDCNetworkIsRemoved
}

@test "driver: docker-compose, action: run, exit status: not 0" {
  cleanUpDCContainers
  cleanUpDCNetwork
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 notexistentcommand
  assert_output --partial "Dojo version"
  assert_output --partial "Current shell is interactive: false"
  # this is the error, if we don't use "init: true" in docker-compose file
  # assert_output --partial "executable file not found"
  assert_output --partial "exec notexistentcommand failed: No such file or directory"
  assert_output --partial "Exit status from run command: 127"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  assert_equal "$status" 127
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
  testDCContainersAreRemoved
  testDCNetworkIsRemoved
}
# # TODO: this is unsupported, #17186
# @test "driver: docker-compose, action: run, command: unset" {
#   cleanUpDCContainers
#   cleanUpDCNetwork
#   cleanUpEnvFiles
#   cleanUpDCDojoFile
#   run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true -i=false --image=alpine:3.8
#   assert_output --partial "Dojo version"
#   refute_output --partial "WARN"
#   refute_output --partial "warn"
#   refute_output --partial "ERROR"
#   refute_output --partial "error"
#   assert_equal "$status" 0
#   testEnvFileIsRemoved
#   testDCDojoFileIsRemoved
#   testDCContainersAreRemoved
#   testDCNetworkIsRemoved
# }
@test "driver: docker-compose, action: run, command: set after --" {
  cleanUpDCContainers
  cleanUpDCNetwork
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 -- whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "whoami"
  assert_output --partial "Exit status from run command: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
  testDCContainersAreRemoved
  testDCNetworkIsRemoved
}
@test "driver: docker-compose, action: run, command with quotes" {
  cleanUpDCContainers
  cleanUpDCNetwork
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 sh -c "echo hello"
  assert_output --partial "Dojo version"
  assert_output --partial "sh -c \"echo hello\""
  assert_output --partial "Exit status from run command: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
  testDCContainersAreRemoved
  testDCNetworkIsRemoved
}
@test "driver: docker-compose, action: run, environment preserved" {
  cleanUpDCContainers
  cleanUpDCNetwork
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run bash -c "export ABC=custom_value ; ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 sh -c 'env | grep ABC'"
  assert_output --partial "Dojo version"
  assert_output --partial "custom_value"
  # set in docker-compose file
  assert_output --partial "1234"
  assert_output --partial "Exit status from run command: 0"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
  testDCContainersAreRemoved
  testDCNetworkIsRemoved
}