# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "dojo --version returns version string" {
  run ${DOJO_PATH} --version
  [[ "$output" =~ 'Dojo version' ]]
  [ "$status" -eq 0 ]
}
@test "dojo --help returns usage info" {
  run ${DOJO_PATH} --help
  [[ "$output" =~ 'Usage of dojo' ]]
  [[ "$output" =~ 'Driver: docker or docker-compose' ]]
  [ "$status" -eq 0 ]
}
