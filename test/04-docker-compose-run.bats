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

@test "driver: docker-compose, action: run, exit status: 0" {
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "whoami"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
}
@test "driver: docker-compose, action: run, exit status: not 0" {
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 notexistentcommand
  assert_output --partial "Dojo version"
  assert_output --partial "Current shell is interactive: false"
  assert_output --partial "executable file not found"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  assert_equal "$status" 1
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
}
# # TODO: this is unsupported, #17186
# @test "driver: docker-compose, action: run, command: unset" {
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
# }
@test "driver: docker-compose, action: run, command: set after --" {
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 -- whoami
  assert_output --partial "Dojo version"
  assert_output --partial "root"
  assert_output --partial "whoami"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
}
@test "driver: docker-compose, action: run, command with quotes" {
  cleanUpEnvFiles
  cleanUpDCDojoFile
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 sh -c "echo hello"
  assert_output --partial "Dojo version"
  assert_output --partial "sh -c \"echo hello\""
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
  testEnvFileIsRemoved
  testDCDojoFileIsRemoved
}
