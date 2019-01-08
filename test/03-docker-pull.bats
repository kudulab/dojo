load '/opt/bats-support/load.bash'
load '/opt/bats-assert/load.bash'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "driver: docker, action: pull, exit status: 0" {
  run ${DOJO_PATH} --action=pull --image=alpine:3.8
  assert_output --partial "Dojo version"
  assert_output --partial "Pulling from library/alpine"
  refute_output --partial "WARN"
  refute_output --partial "warn"
  refute_output --partial "ERROR"
  refute_output --partial "error"
  assert_equal "$status" 0
}