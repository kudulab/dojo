load '/opt/bats-support/load.bash'
load '/opt/bats-assert/load.bash'

DOJO_PATH="../bin/dojo"

# All the dojo scripts are installed, but dojo entrypoint was not run.

@test "/usr/bin/entrypoint.sh file exists and is executable" {
    # the -c below is needed, because our entrypoint is "bash"
    run /bin/bash -c "${DOJO_PATH} --config=Dojofile.to_be_tested_scripts -- -c \"ls -la /usr/bin/entrypoint.sh && test -x /usr/bin/entrypoint.sh\""
    # this is printed on test failure
    echo "output: $output"
    assert_equal "$status" 0
}
@test "/etc/dojo.d, /etc/dojo.d/scripts, /etc/dojo.d/variables drectories exist" {
    run /bin/bash -c "${DOJO_PATH} --config=Dojofile.to_be_tested_scripts -- -c \"test -d /etc/dojo.d && test -d /etc/dojo.d/scripts && test -d /etc/dojo.d/variables\""
    # this is printed on test failure
    echo "output: $output"
    assert_equal "$status" 0
}
@test "/etc/dojo.d/scripts/50-fix-uid-gid.sh file exists and is executable" {
    run /bin/bash -c "${DOJO_PATH} --config=Dojofile.to_be_tested_scripts -- -c \"ls -la /etc/dojo.d/scripts/50-fix-uid-gid.sh && test -x /etc/dojo.d/scripts/50-fix-uid-gid.sh\""
    # this is printed on test failure
    echo "output: $output"
    assert_equal "$status" 0
}
@test "/etc/dojo.d/scripts/90-run-user.sh file exists and is executable" {
    run /bin/bash -c "${DOJO_PATH} --config=Dojofile.to_be_tested_scripts -- -c \"ls -la /etc/dojo.d/scripts/90-run-user.sh && test -x /etc/dojo.d/scripts/90-run-user.sh\""
    # this is printed on test failure
    echo "output: $output"
    assert_equal "$status" 0
}
@test "/etc/dojo.d/scripts/29-not-executable-file.sh is NOT executable" {
    run /bin/bash -c "${DOJO_PATH} --config Dojofile.to_be_tested_scripts -- -c \"test -x /etc/dojo.d/scripts/29-not-executable-file.sh\""
    # this is printed on test failure
    echo "output: $output"
    assert_equal "$status" 1
}

# Let's run the dojo entrypoint.
# (Do not run /etc/dojo.d/scripts/* on their own here,
# because they need variables which are sourced by /usr/bin/entrypoint.sh).
@test "/usr/bin/entrypoint.sh returns 0" {
    run /bin/bash -c "${DOJO_PATH} --config Dojofile.to_be_tested_scripts -- -c \"/usr/bin/entrypoint.sh whoami 2>&1\""
    assert_line --partial "dojo init finished"
    refute_output --partial "root"
    # The file /etc/dojo.d/scripts/29-not-executable-file.sh was non-executable,
    # but dojo entrypoint changed it and ran that file
    assert_line --partial "Running a script which user forgot to make executable"
    assert_equal "$status" 0
}
@test "/usr/bin/entrypoint.sh provides secrets and configuration" {
    run /bin/bash -c "${DOJO_PATH} --config Dojofile.to_be_tested_scripts -- -c \"/usr/bin/entrypoint.sh whoami 2>&1 && stat -c %U /home/dojo/.ssh/id_rsa && cat /home/dojo/.ssh/id_rsa\""
    # this is printed on test failure
    echo "output: $output"
    assert_line --partial "dojo"
    # Custom file: /etc/dojo.d/scripts/30-copy-ssh-configs.sh was run by dojo entrypoint
    assert_line --partial "inside id_rsa"
    refute_output --partial "root"
    assert_equal "$status" 0
}
@test "cleanup" {
    # make the non-executable file, non-executable again, so that tests can be run many times
    chmod 644 test/test-files/etc_dojo.d/scripts/29-not-executable-file.sh
}
