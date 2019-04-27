load 'common'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "driver: docker, action: pull, exit status: 0" {
  run ${DOJO_PATH} --debug=true --action=pull --image=alpine:3.8
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "Pulling from library/alpine" ]]
  [[ "$output" =~ "Exit status from pull command: 0" ]]
  assertNoErrorsOrWarnings
  [ "$status" -eq 0 ]
}
