import os
import subprocess

test_dir = str(os.path.dirname(os.path.abspath(__file__)))
project_root = os.path.join(test_dir, '..', '..')

def decode_utf8(bytes):
    return bytes.decode('utf-8')

def run_command(exe, args, env=None):
    cwd = project_root
    if env is None:
        result = subprocess.run([exe] + args, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=False, cwd=cwd)
    else:
        result = subprocess.run(str.join (' ', [exe] + args), stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True, cwd=cwd, env=env)
    result.stdout = decode_utf8(result.stdout)
    result.stderr = decode_utf8(result.stderr)
    return result

def run_shell(command):
    return subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)

def run_dojo(args, env=None):
    dojo_exe = os.path.join(test_dir, '..', '..', 'bin', 'dojo')
    return run_command(dojo_exe, args, env)

def assert_no_warnings_or_errors(text):
    assert not 'warn' in text
    assert not 'error' in text
    assert not 'WARN' in text
    assert not 'ERROR' in text