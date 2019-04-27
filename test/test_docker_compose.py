import os

from .support.common import *


def cleanUpDCDojoFile():
    try:
        os.remove(os.path.join(project_root, 'test/test-files/itest-dc.yaml.dojo'))
    except FileNotFoundError:
        pass

def testDCDojoFileIsRemoved():
    assert not os.path.exists(os.path.join(project_root, 'test/test-files/itest-dc.yaml.dojo'))

def cleanUpDCContainers():
    run_command('docker', ['stop', 'testdojorunid_default_run_1'])
    run_command('docker', ['stop', 'testdojorunid_abc_1'])
    run_command('docker', ['rm', 'testdojorunid_default_run_1'])
    run_command('docker', ['rm', 'testdojorunid_abc_1'])

def testDCContainersAreRemoved():
    result = run_command('docker', ['ps', '-a', '--filter', 'name=testdojorunid'])
    assert not "testdojorunid" in result.stdout
    assert result.returncode == 0

def cleanUpDCNetwork():
    run_command('docker', ['network', 'rm', 'testdojorunid_default'])

def testDCNetworkIsRemoved():
    result = run_command('docker',  ['network', 'ls', '--filter', "name=testdojorunid"])
    assert not "testdojorunid" in result.stdout
    assert result.returncode == 0


def test_docker_compose_run_when_exit_zero():
    cleanUpDCContainers()
    cleanUpDCNetwork()
    cleanUpDCDojoFile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 whoami".split(' '))
    assert 'Dojo version' in result.stderr
    assert result.returncode == 0
    assert 'root' in result.stdout
    assert 'whoami' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    testDCDojoFileIsRemoved()
    testDCContainersAreRemoved()
    testDCNetworkIsRemoved()

def test_docker_compose_run_command_output_capture():
    cleanUpDCContainers()
    cleanUpDCNetwork()
    cleanUpDCDojoFile()
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', "printenv HOME"])
    assert result.stdout == '/root\n'
    assert "Exit status from run command: 0" in result.stderr
    assert "Exit status from cleaning: 0" in result.stderr
    assert "Exit status from signals: 0" in result.stderr
    assert "Dojo version" in result.stderr

def test_docker_compose_run_when_exit_non_zero():
    cleanUpDCContainers()
    cleanUpDCNetwork()
    cleanUpDCDojoFile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 notexistentcommand".split(' '))
    assert 'Dojo version' in result.stderr
    assert "Current shell is interactive: false" in result.stderr
    assert "exec notexistentcommand failed: No such file or directory" in result.stderr
    assert "Exit status from run command: 127" in result.stderr
    assert 127 == result.returncode
    testDCDojoFileIsRemoved()
    testDCContainersAreRemoved()
    testDCNetworkIsRemoved()

def test_docker_compose_run_when_double_dash_command_split():
    cleanUpDCContainers()
    cleanUpDCNetwork()
    cleanUpDCDojoFile()
    result = run_dojo("--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --test=true --image=alpine:3.8 -- whoami".split())
    assert 'Dojo version' in result.stderr
    assert result.returncode == 0
    assert 'root' in result.stdout
    assert 'whoami' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    testDCDojoFileIsRemoved()
    testDCContainersAreRemoved()
    testDCNetworkIsRemoved()

def test_docker_compose_run_when_shell_command():
    cleanUpDCContainers()
    cleanUpDCNetwork()
    cleanUpDCDojoFile()
    result = run_dojo(['--driver=docker-compose',  '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', 'echo hello'])
    assert 'Dojo version' in result.stderr
    assert 'Exit status from run command: 0' in result.stderr
    assert 'hello' in result.stdout
    assert result.returncode == 0
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
    testDCDojoFileIsRemoved()
    testDCContainersAreRemoved()
    testDCNetworkIsRemoved()

def test_docker_compose_run_preserves_env_vars():
    cleanUpDCContainers()
    cleanUpDCNetwork()
    cleanUpDCDojoFile()
    envs = dict(os.environ)
    envs['ABC'] ='custom_value'
    result = run_dojo(['--driver=docker-compose', '--dcf=./test/test-files/itest-dc.yaml', '--debug=true', '--test=true', '--image=alpine:3.8', 'sh', '-c', 'env | grep ABC'],
                      env=envs)
    assert 'Dojo version' in result.stderr
    assert 'custom_value' in result.stdout
    assert '1234' in result.stdout
    assert 'Exit status from run command: 0' in result.stderr
    assert result.returncode == 0
    testDCDojoFileIsRemoved()
    testDCContainersAreRemoved()
    testDCNetworkIsRemoved()

def test_docker_compose_pull():
    result = run_dojo('--driver=docker-compose --dcf=./test/test-files/itest-dc.yaml --debug=true --action=pull --image=alpine:3.8'.split(' '))
    assert 'Dojo version' in result.stderr
    assert 'pulling' in result.stderr
    assert "Exit status from pull command: 0" in result.stderr
    assert_no_warnings_or_errors(result.stderr)
    assert_no_warnings_or_errors(result.stdout)
