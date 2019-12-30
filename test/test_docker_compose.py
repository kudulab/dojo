import os
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
    assert not "testdojorunid" in result.stdout
    assert result.returncode == 0


def clean_up_dc_network():
    run_command('docker', ['network', 'rm', 'testdojorunid_default'])


def test_dc_network_is_removed():
    result = run_command('docker',  ['network', 'ls', '--filter', "name=testdojorunid"])
    assert not "testdojorunid" in result.stdout
    assert result.returncode == 0


def test_docker_compose_run_when_exit_zero():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 whoami".split(' '))
    assert 'Dojo version' in result.stderr
    assert result.returncode == 0
    assert 'root' in result.stdout
    assert 'whoami' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_command_output_capture():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', "printenv HOME"])
    assert result.stdout == '/root\n'
    assert "Exit status from run command: 0" in result.stderr
    assert "Exit status from cleaning: 0" in result.stderr
    assert "Exit status from signals: 0" in result.stderr
    assert "Dojo version" in result.stderr


def test_docker_compose_run_when_exit_non_zero():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 notexistentcommand".split(' '))
    assert 'Dojo version' in result.stderr
    assert "Current shell is interactive: false" in result.stderr
    assert "exec notexistentcommand failed: No such file or directory" in result.stderr
    assert "Exit status from run command: 127" in result.stderr
    assert 127 == result.returncode
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_when_double_dash_command_split():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 -- whoami".split())
    assert 'Dojo version' in result.stderr
    assert result.returncode == 0
    assert 'root' in result.stdout
    assert 'whoami' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_when_shell_command():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    result = run_dojo(['--driver=docker-compose',  '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', 'echo hello'])
    assert 'Dojo version' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert 'hello' in result.stdout
    assert result.returncode == 0
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_preserves_env_vars():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    envs = dict(os.environ)
    envs['ABC'] ='custom_value'
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', 'env | grep ABC'],
                      env=envs)
    assert 'Dojo version' in result.stderr
    assert 'custom_value' in result.stdout
    assert '1234' in result.stdout
    assert 'Exit status from run command: 0' in result.stderr
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
        '--image=alpine:3.8', 'sh', '-c', '"source /etc/dojo.d/variables/00-multiline-vars.sh && env | grep -A 1 ABC"'],
        env=envs)
    assert 'Dojo version' in result.stderr
    assert '/etc/dojo.d/variables/00-multiline-vars.sh' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    assert result.returncode == 0
    assert 'Exit status from run command:' in result.stderr
    assert """first line
second line""" in result.stdout
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_pull():
    result = run_dojo('--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --action=pull --image=alpine:3.8'.split(' '))
    assert 'Dojo version' in result.stderr
    assert 'pulling' in result.stderr
    assert "Exit status from pull command: 0" in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)


def test_docker_compose_dojo_work_variables():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    os.makedirs(os.path.join(project_root, 'test/test-files/custom-dir-env-var'), exist_ok=True)
    with open(os.path.join(project_root, 'test/test-files/custom-dir-env-var/file1.txt'), 'w') as f:
        f.write('123')
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc-env-var.yaml',
                       '--debug=true', '--test=true', '--image=alpine:3.8', '--', 'sh',
                       '-c', "cat /dojo/work/custom-dir/file1.txt"])
    assert "Dojo version" in result.stderr
    assert not "DOJO_WORK_OUTER variable is not set" in result.stderr
    assert not "DOJO_WORK_INNER variable is not set" in result.stderr
    assert '123' in result.stdout
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    assert result.returncode == 0
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_shows_nondefault_containers_logs_when_all_constainers_succeeded():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    # make the command of the default container last long enough so that the other
    # container is started and managed to produce some output
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc-verbose.yaml',
                       '--print-logs=always',
                       '--debug=true', '--test=true', '--image=alpine:3.8', '--', 'sh',
                       '-c', "echo 1; sleep 1; echo 2; sleep 1;"])
    assert 'Dojo version' in result.stderr
    assert result.returncode == 0
    assert 'echo 1; sleep 1; echo 2; sleep 1;' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert 'Here are logs of container: testdojorunid_abc_1' in result.stderr
    assert 'iteration: 1' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
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
                       '--debug=true', '--test=true', '--image=alpine:3.8', '--', 'sh',
                       '-c', "echo 1; sleep 1; echo 2; sleep 1;"])
    assert 'Dojo version' in result.stderr
    assert result.returncode == 0
    assert 'echo 1; sleep 1; echo 2; sleep 1;' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert 'Here are logs of container: testdojorunid_abc_1' in result.stderr
    assert 'some-non-existent-command: not found' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()


def test_docker_compose_run_shows_nondefault_containers_logs_when_default_container_failed():
    clean_up_dc_containers()
    clean_up_dc_network()
    clean_up_dc_dojofile()
    # make the command of the default container last long enough so that the other
    # container is started and managed to produce some output
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc-verbose.yaml --print-logs=failure --debug=true --test=true --image=alpine:3.8 -- some-non-existent-command".split())
    assert 'Dojo version' in result.stderr
    assert result.returncode == 127
    assert 'Exit status from run command: 127' in result.stderr
    assert 'Here are logs of container: testdojorunid_abc_1' in result.stderr
    assert 'iteration: 1' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    test_dc_dojofile_is_removed()
    test_dc_containers_are_removed()
    test_dc_network_is_removed()