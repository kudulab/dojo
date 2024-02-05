import pytest
from .support.common import run_dojo, run_command, get_dojo_exe


def test_version_output():
    result = run_dojo(['--version'])
    assert 'Dojo version' in result.stdout
    assert result.returncode == 0


def test_help_output():
    result = run_dojo(['--help'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Usage of dojo' in result.stderr, dojo_combined_output_str
    assert 'Driver: docker or docker-compose' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0


def test_dojo_prints_error_if_bash_not_installed():
    dojo_exe = get_dojo_exe()
    result = run_command('docker', ['run', '-t', '--rm', '-v', f"{dojo_exe}:/usr/bin/dojo", "alpine:3.19", "/usr/bin/dojo"])
    combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Error while verifying if Bash is installed' in result.stdout, combined_output_str
    assert result.returncode == 1
