
function cleanUpEnvFiles() {
  rm -f /tmp/test-dojo*
}

function testEnvFileIsRemoved(){
  run bash -c "ls -la /tmp/ | grep 'test-dojo'"
  assert_equal "$status" 1
}