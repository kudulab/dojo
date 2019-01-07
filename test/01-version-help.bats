load '/opt/bats-support/load.bash'
load '/opt/bats-assert/load.bash'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "dojo --version returns version" {
  run ${DOJO_PATH} --version
  assert_output --partial "Dojo version"
  assert_equal "$status" 0
}
@test "dojo --help returns usage info" {
  run ${DOJO_PATH} --help
  assert_output --partial "Usage of dojo"
  assert_output --partial "Driver: docker or docker-compose"
  assert_equal "$status" 0
}