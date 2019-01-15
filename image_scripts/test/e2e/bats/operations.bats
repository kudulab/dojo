load '/opt/bats-support/load.bash'
load '/opt/bats-assert/load.bash'

DOJO_PATH="../bin/dojo"

@test "/usr/bin/entrypoint.sh returns 0" {
  run /bin/bash -c "${DOJO_PATH} --config=Dojofile.to_be_tested \"pwd && whoami\""
  # this is printed on test failure
  echo "output: $output"
  assert_line --partial "dojo init finished"
  assert_line --partial "/dojo/work"
  refute_output --partial "root"
  assert_equal "$status" 0
}
@test "bash is installed" {
  run /bin/bash -c "${DOJO_PATH} --config=Dojofile.to_be_tested \"bash --version\""
  # this is printed on test failure
  echo "output: $output"
  assert_line --partial "GNU bash"
  assert_equal "$status" 0
}
@test "environment is correctly set" {
  run /bin/bash -c "export ABC=custom ; ${DOJO_PATH} --config=Dojofile.to_be_tested \"env\""
  # this is printed on test failure
  echo "output: $output"
  assert_line --partial "dojo_work=/dojo/work"
  assert_line --partial "ABC=custom"
  assert_line --partial "DOJO_USER"
  assert_line --partial "USER=dojo"
  assert_equal "$status" 0
}