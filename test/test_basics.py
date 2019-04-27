import pytest

from .support.common import run_dojo

def test_version_output():
    result = run_dojo(['--version'])
    assert 'Dojo version' in result.stderr #TODO: (ewa) should be in stdout
    assert result.returncode == 0

def test_help_output():
    result = run_dojo(['--help'])
    assert 'Usage of dojo' in result.stderr
    assert 'Driver: docker or docker-compose' in result.stderr
    assert result.returncode == 0