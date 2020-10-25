import pytest
from .support.common import run_dojo


def test_version_output():
    result = run_dojo(['--version'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Dojo version' in result.stderr, dojo_combined_output_str #TODO: (ewa) should be in stdout
    assert result.returncode == 0


def test_help_output():
    result = run_dojo(['--help'])
    dojo_combined_output_str =  "stdout:\n{0}\nstderror:\n{1}".format(result.stdout, result.stderr)
    assert 'Usage of dojo' in result.stderr, dojo_combined_output_str
    assert 'Driver: docker or docker-compose' in result.stderr, dojo_combined_output_str
    assert result.returncode == 0