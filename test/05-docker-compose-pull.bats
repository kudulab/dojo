load '/opt/bats-support/load.bash'
load '/opt/bats-assert/load.bash'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "driver: docker-compose, action: pull, exit status: 0" {
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --action=pull --image=alpine:3.8
  assert_output --partial "Dojo version"
  assert_output --partial "Pulling"
  assert_output --partial "Exit status from pull command: 0"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
}