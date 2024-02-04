import os
import os.path
from .support.common import *


def clean_up_dc_dojofile():
    try:
        os.remove(os.path.join(project_root, 'test/test-files/itest-dc.yaml.dojo'))
    except FileNotFoundError:
        pass


def test_dc_dojofile_is_removed():
    assert not os.path.exists(os.path.join(project_root, 'test/test-files/itest-dc.yaml.dojo'))


def clean_up_dc_containers():
    run_command('docker', ['stop', 'testdojorunid_default_run_1'])
    run_command('docker', ['stop', 'testdojorunid_abc_1'])
    run_command('docker', ['rm', 'testdojorunid_default_run_1'])
    run_command('docker', ['rm', 'testdojorunid_abc_1'])


def test_dc_containers_are_removed():
    result = run_command('docker', ['ps', '-a', '--filter', 'name=testdojorunid'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert not "testdojorunid" in result.stdout, dojo_combined_output_str
    assert result.returncode == 0


def clean_up_dc_network():
    run_command('docker', ['network', 'rm', 'testdojorunid_default'])


def test_dc_network_is_removed():
    result = run_command('docker',  ['network', 'ls', '--filter', "name=testdojorunid"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert not "testdojorunid" in result.stdout, dojo_combined_output_str
    assert result.returncode == 0


def test_docker_compose_run_when_exit_zero():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.19 whoami".split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    assert 'root' in result.stdout, dojo_combined_output_str
    assert 'whoami' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_command_output_capture():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.19', 'sh', '-c', "printenv HOME"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert result.stdout == '/root\n', dojo_combined_output_str
    assert "Exit status from run command: 0" in result.stderr, dojo_combined_output_str
    assert "Exit status from cleaning: 0" in result.stderr, dojo_combined_output_str
    assert "Exit status from signals: 0" in result.stderr, dojo_combined_output_str
    assert "Dojo version" in result.stderr


def test_docker_compose_run_when_exit_non_zero():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.19 notexistentcommand".split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert "Current shell is interactive: false" in result.stderr, dojo_combined_output_str
    assert "exec notexistentcommand failed: No such file or directory" in result.stderr, dojo_combined_output_str
    assert "Exit status from run command: 127" in result.stderr, dojo_combined_output_str
    assert 127 == result.returncode
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_when_double_dash_command_split():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.19 -- whoami".split())
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    assert 'root' in result.stdout, dojo_combined_output_str
    assert 'whoami' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_when_shell_command():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo(['--driver=docker-compose',  '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.19', 'sh', '-c', 'echo hello'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'hello' in result.stdout, dojo_combined_output_str
    assert result.returncode == 0
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_preserves_env_vars():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    envs = dict(os.environ)
    envs['ABC'] ='custom_value'
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.19', 'sh', '-c', 'env | grep ABC'],
                      env=envs)
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'custom_value' in result.stdout, dojo_combined_output_str
    assert '1234' in result.stdout, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_preserves_multiline_env_vars():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    envs = dict(os.environ)
    envs['ABC'] = """first line
second line"""
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true',
        '--image=alpine:3.19', 'sh', '-c', '"source /etc/dojo.d/variables/00-multiline-vars.sh && env | grep -A 1 ABC"'],
        env=envs)
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert '/etc/dojo.d/variables/00-multiline-vars.sh' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    assert 'Exit status from run command:' in result.stderr, dojo_combined_output_str
    assert """first line
second line""" in result.stdout
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


# see also: test_docker_preserves_bash_functions_from_env_vars for more comments
def test_docker_compose_run_preserves_bash_functions():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    envs = dict(os.environ)
    proc = run_dojo_and_set_bash_func(
        ['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true',
                 '--image=alpine:3.19', 'sh', '-c',
                 '"apk add -U bash && bash -c \'source /etc/dojo.d/variables/01-bash-functions.sh && my_bash_func\'"'],
        env=envs)
    stdout_value_bytes, stderr_value_bytes = proc.communicate()
    stdout = str(stdout_value_bytes)
    stderr = str(stderr_value_bytes)
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(stdout, stderr)
    assert 'Dojo version' in stderr, dojo_combined_output_str
    assert 'Written file /tmp/test-dojo-environment-bash-functions-testdojorunid, contents:' in stderr, dojo_combined_output_str
    assert 'my_bash_func() {  echo "hello"' in stderr, dojo_combined_output_str
    assert '/etc/dojo.d/variables/01-bash-functions.sh' in stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(stdout, dojo_combined_output_str)
    # the bash function was invoked
    assert 'hello' in stdout, dojo_combined_output_str
    assert 'Exit status from run command: 0' in stderr, dojo_combined_output_str
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_pull():
    result = run_dojo('--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --action=pull --image=alpine:3.19'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'Pulling' in result.stderr, dojo_combined_output_str
    assert "Exit status from pull command: 0" in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)


def test_docker_compose_pull_when_no_such_image_exists():
    result = run_dojo('--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --action=pull --image=no_such_image91291925129q783187314218194:abc111aaa.9981412'.split(' '))
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert 'repository does not exist or may require \'docker login\'' in result.stderr, dojo_combined_output_str
    assert "Exit status from pull command: 1" in result.stderr, dojo_combined_output_str
    assert "" == result.stdout, dojo_combined_output_str
    assert result.returncode == 18


def test_docker_compose_dojo_work_variables():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    os.makedirs(os.path.join(project_root, 'test/test-files/custom-dir-env-var'), exist_ok=True)
    with open(os.path.join(project_root, 'test/test-files/custom-dir-env-var/file1.txt'), 'w') as f:
        f.write('123')
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc-env-var.yaml',
                       '--debug=true', '--test=true', '--image=alpine:3.19', '--', 'sh',
                       '-c', "cat /dojo/work/custom-dir/file1.txt"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert "Dojo version" in result.stderr, dojo_combined_output_str
    assert not "DOJO_WORK_OUTER variable is not set" in result.stderr, dojo_combined_output_str
    assert not "DOJO_WORK_INNER variable is not set" in result.stderr, dojo_combined_output_str
    assert '123' in result.stdout, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    assert result.returncode == 0
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_shows_nondefault_containers_logs_when_all_containers_succeeded():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    # make the command of the default container last long enough so that the other
    # container is started and managed to produce some output
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc-verbose.yaml',
                       '--print-logs=always',
                       '--debug=true', '--test=true', '--image=alpine:3.19', '--', 'sh',
                       '-c', "echo 1; sleep 1; echo 2; sleep 1;"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    assert 'echo 1; sleep 1; echo 2; sleep 1;' in result.stderr, dojo_combined_output_str
    assert "Containers created" in result.stderr
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    # Docker-compose >2 names the containers using dashes instead of underscores (while underscores
    # were used by Docker-compose <2), so we cannot test for containers' names if we want to
    # support both Docker-compose versions (<2 and >2)
    assert 'Here are logs of container: ' in result.stderr, dojo_combined_output_str
    assert 'which status is: running' in result.stderr, dojo_combined_output_str
    assert 'iteration: 1' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()

# This test reproduces the panic that used to happen in Dojo 0.11.0 when using Docker-compose 2.24.5
# when the default container stopped or was already removed
def test_docker_compose_run_shows_nondefault_containers_logs_when_all_containers_succeeded_while_no_sleep():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    # make the command of the default container last long enough so that the other
    # container is started and managed to produce some output
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc-verbose.yaml',
                       '--print-logs=always',
                       '--debug=true', '--test=true', '--image=alpine:3.19', '--', 'sh',
                       '-c', "echo 1;"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    assert 'echo 1;' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Here are logs of container: ' in result.stderr, dojo_combined_output_str
    assert 'which status is: running' in result.stderr, dojo_combined_output_str
    assert 'iteration: 1' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_shows_nondefault_containers_logs_when_nondefault_container_failed():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    # make the command of the default container last long enough so that the other
    # container is started and managed to produce some output
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc-verbose-fail.yaml',
                       '--print-logs=always',
                       '--debug=true', '--test=true', '--image=alpine:3.19', '--', 'sh',
                       '-c', "echo 1; sleep 1; echo 2; sleep 1;"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    assert 'echo 1; sleep 1; echo 2; sleep 1;' in result.stderr, dojo_combined_output_str
    assert 'Exit status from run command: 0' in result.stderr, dojo_combined_output_str
    assert 'Here are logs of container: ' in result.stderr, dojo_combined_output_str
    assert 'which exited with exitcode: 127' in result.stderr, dojo_combined_output_str
    assert 'some-non-existent-command: not found' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_shows_nondefault_containers_logs_when_default_container_failed():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    # make the command of the default container last long enough so that the other
    # container is started and managed to produce some output
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc-verbose.yaml --print-logs=failure --debug=true --test=true --image=alpine:3.19 -- some-non-existent-command".split())
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert result.returncode == 127
    assert 'Exit status from run command: 127' in result.stderr, dojo_combined_output_str
    assert 'Here are logs of container: ' in result.stderr, dojo_combined_output_str
    assert 'which status is: running' in result.stderr, dojo_combined_output_str
    assert 'iteration: 1' in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def clean_up_dojo_logs_file(logs_file):
    try:
        os.remove(os.path.join(project_root, logs_file))
    except FileNotFoundError:
        pass


def test_docker_compose_run_shows_nondefault_containers_logs_when_all_constainers_succeeded_print_logs_to_file():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    logs_file_dcver1 = "dojo-logs-testdojorunid_abc_1-testdojorunid.txt"
    logs_file_dcver2 = "dojo-logs-testdojorunid-abc-1-testdojorunid.txt"
    clean_up_dojo_logs_file(logs_file_dcver1)
    clean_up_dojo_logs_file(logs_file_dcver2)

    # make the command of the default container last long enough so that the other
    # container is started and managed to produce some output
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc-verbose.yaml',
                       '--print-logs=always', '--print-logs-target=file',
                       '--debug=false', '--test=true', '--image=alpine:3.19', '--', 'sh',
                       '-c', "echo 1; sleep 1; echo 2; sleep 1;"])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0
    assert 'echo 1; sleep 1; echo 2; sleep 1;' in result.stderr, dojo_combined_output_str
    # Docker-compose >2 names the containers using dashes instead of underscores (while underscores
    # were used by Docker-compose <2), so we cannot test for containers' names if we want to
    # support both Docker-compose versions (<2 and >2)
    assert 'The logs of container:' in result.stderr, dojo_combined_output_str
    assert 'were saved to file: ' in result.stderr, dojo_combined_output_str
    assert 'testdojorunid.txt' in result.stderr, dojo_combined_output_str
    if os.path.isfile(logs_file_dcver1):
        with open(logs_file_dcver1, "r") as file:
            contents = file.readlines()
            assert 'iteration: 1\n' in contents
            assert 'stdout:\n' in contents
            assert 'stderr:\n' in contents
    if os.path.isfile(logs_file_dcver2):
        with open(logs_file_dcver2, "r") as file:
            contents = file.readlines()
            assert 'iteration: 1\n' in contents
            assert 'stdout:\n' in contents
            assert 'stderr:\n' in contents
    # exactly one of these files should exist (depending on which docker-compose version we run)
    assert (os.path.isfile(logs_file_dcver1) or os.path.isfile(logs_file_dcver2)), True

    assert 'iteration: 1' not in result.stderr, dojo_combined_output_str
    assert_no_warnings_or_errors(result.stderr, dojo_combined_output_str)
    assert_no_warnings_or_errors(result.stdout, dojo_combined_output_str)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()
    clean_up_dojo_logs_file(logs_file_dcver1)
    clean_up_dojo_logs_file(logs_file_dcver2)



def test_docker_compose_run_in_directory_with_capital_letters():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()

    dojo_exe = os.path.join(test_dir, '..', '..', 'bin', 'dojo')
    dojo_exe_absolute_path = os.path.abspath(dojo_exe)
    dojo_args = ['--driver=docker-compose', '--dcf=./itest-dc.yaml', '--debug=true', '--image=alpine:3.19', 'sh', '-c', "printenv HOME"]
    result = subprocess.run([dojo_exe] + dojo_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=False, cwd='test/test-files/DirWithUpperCaseLetters')
    result.stdout = decode_utf8(result.stdout)
    result.stderr = decode_utf8(result.stderr)

    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert result.stdout == '/root\n', dojo_combined_output_str
    assert "Exit status from run command: 0" in result.stderr, dojo_combined_output_str
    assert "Exit status from cleaning: 0" in result.stderr, dojo_combined_output_str
    assert "Exit status from signals: 0" in result.stderr, dojo_combined_output_str
    assert "Dojo version" in result.stderr
