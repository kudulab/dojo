from test.support.common import *


def clean_up_docker_container():
    run_shell('docker ps -a -q --filter "name=testdojorunid" | xargs --no-run-if-empty docker stop')
    run_shell('docker ps -a -q --filter "name=testdojorunid" | xargs --no-run-if-empty docker rm')


def test_docker_container_is_removed():
    result = run_command('docker', ['ps', '-a', '--filter', "name=testdojorunid"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert not 'testdojorunid' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0


def test_docker_when_zero_exit():
    clean_up_docker_container()
    result = run_dojo('--debug=true --test=true --image=alpine:3.8 whoami'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'root' in result.stdout, dojo_combined_output_str
    assert 'alpine:3.8 whoami' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    test_docker_container_is_removed()


def test_docker_capture_output():
    clean_up_docker_container()
    # run this one test manually: pytest --capture=fd  --verbose test/test_docker.py::test_docker_capture_output
    # run this without pytest: ./bin/dojo --debug=true --test=true --image=alpine:3.8 sh -c "printenv HOME"
    result = run_dojo(['--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', "printenv HOME"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert '/root\n' == result.stdout, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    test_docker_container_is_removed()


def test_docker_capture_output_when_unable_to_pull_image():
    clean_up_docker_container()
    # pytest --capture=fd  --verbose test/test_docker.py::test_docker_capture_output_when_unable_to_pull_image
    # ./bin/dojo --debug=true --test=true --image=alpine:3.8 sh -c "printenv HOME && hostname"
    result = run_dojo(['--debug=true', '--test=true', '--image=no_such_image91291925129q783187314218194:abc111aaa.9981412', 'sh', '-c', "printenv HOME"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    # this is the main reason for this test:
    assert 'Unable to find image' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 125' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert result.returncode == 125
    test_docker_container_is_removed()


def test_docker_when_non_existent_command():
    clean_up_docker_container()
    result = run_dojo('--debug=true --test=true --image=alpine:3.8 notexistentcommand'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'executable file not found' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 127' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 127
    test_docker_container_is_removed()


def test_docker_when_no_command():
    clean_up_docker_container()
    result = run_dojo('--debug=true --test=true --image=alpine:3.8 -i=false'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    test_docker_container_is_removed()


def test_docker_when_double_dash_command_split():
    clean_up_docker_container()
    result = run_dojo('--debug=true --test=true --image=alpine:3.8 -- whoami'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'root' in result.stdout, dojo_combined_output_str
    assert 'alpine:3.8 whoami' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    test_docker_container_is_removed()


def test_docker_when_shell_command():
    clean_up_docker_container()
    result = run_dojo(['--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', 'echo hello'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'hello' in result.stdout, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    test_docker_container_is_removed()


def test_docker_preserves_env_vars():
    clean_up_docker_container()
    envs = dict(os.environ)
    envs['ABC'] = 'custom_value'
    result = run_dojo(
        ['--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', 'env | grep ABC'],
        env=envs)
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'custom_value' in result.stdout, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0

# Bash experiments:
# $ export MULTILINE1="first line\nsecond line"
# $ export MULTILINE2="first line
# > second line"
# $ read -d '' MULTILINE3 <<EOF
# > first line
# > second line
# > EOF
# $ export MULTILINE3
#
# $ echo $MULTILINE1
# first line\nsecond line
# $ echo $MULTILINE2
# first line second line
# $ echo $MULTILINE3
# first line second line
# $ echo "$MULTILINE1"
# first line\nsecond line
# $ echo "$MULTILINE2"
# first line
# second line
# $ echo "$MULTILINE3"
# first line
# second line
#
# In Dojo, only the MULTILINE2 and MULTILINE3 will be put among multiline variables.
# MULTILINE1 will be treated as oneline variable.

def test_docker_preserves_multiline_env_vars():
    clean_up_docker_container()
    envs = dict(os.environ)
    envs['ABC'] = """first line
second line"""
    result = run_dojo(
        # We need to source the file: /etc/dojo.d/variables/01-bash-functions.sh
        # explicitly, because the alpine docker image is not a Dojo image, i.e.
        # it does not have the Dojo entrypoint.sh.
        ['--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', '"source /etc/dojo.d/variables/00-multiline-vars.sh && env | grep -A 1 ABC"'],
        env=envs)
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert '/etc/dojo.d/variables/00-multiline-vars.sh' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    assert 'Exit status from run command:' in result.stderr, dojo_combined_output_str
    assert """first line
second line""" in result.stdout, dojo_combined_output_str


def test_docker_preserves_bash_functions_from_env_vars():
    clean_up_docker_container()
    envs = dict(os.environ)
    # the following does not influence the dojo process
    # envs['BASH_FUNC_my_bash_func%%'] = """()) {  echo "hello"
# }"""
    proc = run_dojo_and_set_bash_func(
        # We need to source the file: /etc/dojo.d/variables/01-bash-functions.sh
        # explicitly, because the alpine docker image is not a Dojo image, i.e.
        # it does not have the Dojo entrypoint.sh. Even if it had,
        # we'd still have to source the file explicitly, because
        # sudo does not preserve bash functions.
        # Dojo entrypoint sources this file too, but then it runs sudo.
        # https://unix.stackexchange.com/questions/549140/why-doesnt-sudo-e-preserve-the-function-environment-variables-exported-by-ex
        # https://unix.stackexchange.com/a/233097
        ['--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', '"apk add -U bash && bash -c \'source /etc/dojo.d/variables/01-bash-functions.sh && my_bash_func\'"'],
        env=envs)
    stdout_value_bytes, stderr_value_bytes = proc.communicate()
    stdout = stdout_value_bytes.decode("utf-8")
    stderr = stderr_value_bytes.decode("utf-8")
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(stdout, stderr)
    assert 'Dojo version' in stderr, dojo_combined_output_str
    # print(stdout)
    # print(stderr)
    assert 'Written file /tmp/test-dojo-environment-bash-functions-testdojorunid, contents:' in stderr, dojo_combined_output_str
    assert 'my_bash_func() {  echo "hello"' in stderr, dojo_combined_output_str
    assert '/etc/dojo.d/variables/01-bash-functions.sh' in stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(stdout, dojo_combined_output_str)
    # the bash function was invoked
    assert 'hello' in stdout, dojo_combined_output_str
    assert 'Exit status from run command: 0' in stderr, dojo_combined_output_str


def test_docker_when_custom_relative_directory():
    clean_up_docker_container()
    result = run_dojo(['-c', 'test/test-files/Dojofile', '--debug=true', '--test=true', '--image=alpine:3.8', 'whoami'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'root' in result.stdout, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    test_docker_container_is_removed()


def test_docker_when_nonexistent_custom_relative_directory():
    clean_up_docker_container()
    try:
        os.removedirs(os.path.join(project_root, 'test/test-files/not-existent'))
    except FileNotFoundError:
        pass
    result = run_dojo(['-c', 'test/test-files/Dojofile.work_not_exists', '--debug=true', '--test=true', '--image=alpine:3.8', 'whoami'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'root' in result.stdout, dojo_combined_output_str
    assert "test/test-files/not-existent does not exist" in result.stderr, dojo_combined_output_str
    assert 'WARN' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from cleaning: 0' in result.stderr, dojo_combined_output_str
    assert 'Exit status from signals: 0' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    test_docker_container_is_removed()


def test_docker_pull_when_image_can_be_pulled():
    result = run_dojo('--debug=true --action=pull --image=alpine:3.8'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert "Dojo version" in result.stderr, dojo_combined_output_str
    assert "Pulling from library/alpine" in result.stdout, dojo_combined_output_str
    assert result.returncode == 0
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)


def test_docker_pull_when_no_such_image_exists():
    result = run_dojo('--debug=true --action=pull --image=no_such_image91291925129q783187314218194:abc111aaa.9981412'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert "" == result.stdout, dojo_combined_output_str
    assert "Dojo version" in result.stderr, dojo_combined_output_str
    assert "repository does not exist or may require 'docker login'" in result.stderr, dojo_combined_output_str
    assert result.returncode == 1