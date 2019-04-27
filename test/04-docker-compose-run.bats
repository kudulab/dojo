load 'common'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

function cleanUpDCDojoFile() {
  rm -f test/test-files/itest-dc.yaml.dojo
}

function testDCDojoFileIsRemoved(){
  run test -f test/test-files/itest-dc.yaml.dojo
  [ "$status" -eq 1 ]
}

function cleanUpDCContainers() {
  docker stop testdojorunid_default_run_1 >/dev/null 2>&1 || true
  docker stop testdojorunid_abc_1 >/dev/null 2>&1 || true
  docker rm testdojorunid_default_run_1 >/dev/null 2>&1 || true
  docker rm testdojorunid_abc_1 >/dev/null 2>&1 || true
}

function testDCContainersAreRemoved() {
  run docker ps -a --filter "name=testdojorunid"
  [[ ! "$output" =~ "testdojorunid" ]]
  [ "$status" -eq 0 ]
}

function cleanUpDCNetwork() {
  docker network rm testdojorunid_default || true
}

function testDCNetworkIsRemoved() {
  run docker network ls --filter "name=testdojorunid"
  [[ ! "$output" =~ "testdojorunid" ]]
  [ "$status" -eq 0 ]
}

@test "driver: docker-compose, action: run, exit status: 0" {
  cleanUpDCContainers
  cleanUpDCNetwork
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 whoami
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "root" ]]
  [[ "$output" =~ "whoami" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
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
  [ "$status" -eq 0 ]
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
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "Current shell is interactive: false" ]]
  # this is the error, if we don't use "init: true" in docker-compose file
  # [[ "$output" =~ "executable file not found"
  [[ "$output" =~ "exec notexistentcommand failed: No such file or directory" ]]
  [[ "$output" =~ "Exit status from run command: 127" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
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
#   [[ "$output" =~ "Dojo version"
#   [[ ! "$output" =~ "WARN"
#   [[ ! "$output" =~ "warn"
#   [[ ! "$output" =~ "ERROR"
#   [[ ! "$output" =~ "error"
#   [ "$status" -eq 0 ]
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
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "root" ]]
  [[ "$output" =~ "whoami" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
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
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "sh -c \"echo hello\"" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [[ ! "$output" =~ "WARN" ]]
  [[ ! "$output" =~ "warn" ]]
  [[ ! "$output" =~ "ERROR" ]]
  [[ ! "$output" =~ "error" ]]
  [ "$status" -eq 0 ]
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
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "custom_value" ]]
  # set in docker-compose file
  [[ "$output" =~ "1234" ]]
  [[ "$output" =~ "Exit status from run command: 0" ]]
  [ "$status" -eq 0 ]
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
  testDCContainersAreRemoved
  testDCNetworkIsRemoved
}
