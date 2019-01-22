
function cleanUpEnvFiles() {
  rm -f /tmp/test-dojo*
}

function testEnvFileIsRemoved(){
  run bash -c "ls -la /tmp/ | grep 'test-dojo'"
  assert_equal "$status" 1
}

function cleanUpDockerContainer() {
  docker ps -a -q --filter "name=testdojorunid" | xargs --no-run-if-empty docker stop
  docker ps -a -q --filter "name=testdojorunid" | xargs --no-run-if-empty docker rm
}

function testDockerContainerIsRemoved(){
  run docker ps -a --filter "name=testdojorunid"
  refute_output --partial "testdojorunid"
  assert_equal "$status" 0
}