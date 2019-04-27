
function cleanUpEnvFiles() {
  rm -f /tmp/test-dojo*
}

function testEnvFileIsRemoved(){
  run bash -c "ls -la /tmp/ | grep 'test-dojo'"
  [ "$status" -eq 1 ]
}

function cleanUpDockerContainer() {
  docker ps -a -q --filter "name=testdojorunid" | xargs --no-run-if-empty docker stop
  docker ps -a -q --filter "name=testdojorunid" | xargs --no-run-if-empty docker rm
}

function testDockerContainerIsRemoved(){
  run docker ps -a --filter "name=testdojorunid"
  [[ ! "$output" =~ 'testdojorunid' ]]
  [ "$status" -eq 0 ]
}

function assertNoErrorsOrWarnings {
  [[ ! "$output" =~ 'WARN' ]]
  [[ ! "$output" =~ 'warn' ]]
  [[ ! "$output" =~ 'ERROR' ]]
  [[ ! "$output" =~ 'error' ]]
}
