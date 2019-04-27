load 'common'

# make absolute path out of relative
DOJO_PATH=$(readlink -f "./bin/dojo")

@test "driver: docker-compose, action: pull, exit status: 0" {
  run ${DOJO_PATH} --driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --action=pull --image=alpine:3.8
  [[ "$output" =~ "Dojo version" ]]
  [[ "$output" =~ "Pulling" ]]
  [[ "$output" =~ "Exit status from pull command: 0" ]]
  assertNoErrorsOrWarnings
  [ "$status" -eq 0 ]
}
