#!/bin/bash

test_file="dojo-signal-test-output"

function log_test() {
  echo "Signal test: $1"
}

function run_test_process_and_send_signal() {
    local test_process=${1?test_process not set}
    local kill_after_this_string=${2?kill_after_this_string not set}
    log_test "Will run test process: ${test_process}"

    # Start job in the background writing to a new file
    test_process_exit_status=0
    output=""
    (set -x; rm -f "${test_file}"; touch "${test_file}"; )
    ${test_process} >"${test_file}" 2>"${test_file}" &
    local pid=$!
    # local pgid=$(ps -o pgid ${pid} | tail -1 | tr -d " ")
    log_test "Initiated test process with PID: ${pid}, PGID: ${pgid}"

    while true; do
       if [[ $(grep "command will be" -c ${test_file}) == 1 ]] && [[ $(grep "${kill_after_this_string}" -c ${test_file}) == 4 ]]; then
            log_test "Can kill the process now"
            break;
       fi
       test_process_output=$(cat "${test_file}")
       if [[ $(echo $test_process_output | wc -l) == "1" ]] && [[ "${test_process_output}" == *"No such file or directory"* ]]; then
            log_test "Error while running the test process: ${test_process_output}"
            exit 1;
       fi
       log_test "Waiting 1s, if string appears: \"${kill_after_this_string}\" will kill the test process"
       sleep 1
    done

    # Kill the dojo process. If we kill the whole process group, no cleanup is done.
    ( set -x; kill -15 ${pid} && echo "successfully killed" || echo "failed to kill"; )
    log_test "Waiting for the test process to be finished"
    wait ${pid}
    test_process_exit_status=$?
    log_test "The test process is finished"

    output=$(cat "${test_file}")
    log_test "Output from test process:"
    log_test "-----------------------------"
    echo "${output}"
    log_test "-----------------------------"
    log_test "Exit status: ${test_process_exit_status}"
}

function test_value() {
  actual=$1
  expected=$2
  if [[ "${actual}" != "${expected}" ]]; then
    log_test "Expected: ${expected}, got: ${actual}" >&2
    return 1
  fi
  return 0
}
function string_contains() {
  my_string=$1
  contained_string=$2
  if [[ "${my_string}" != *"${contained_string}"* ]]; then
    log_test "Expected this string to be contained, but was not: ${contained_string}" >&2
    return 1
  else
    log_test "Expected this string to be contained, and it was: ${contained_string}" >&2
    return 0
  fi
}
function string_not_contains() {
  my_string=$1
  contained_string=$2
  if [[ "${my_string}" == *"${contained_string}"* ]]; then
    log_test "Expected this string NOT to be contained and it was: ${contained_string}" >&2
    return 1
  else
    log_test "Expected this string NOT to be contained, but it was: ${contained_string}" >&2
    return 0
  fi
}

# Test 1.: driver=docker, container's entrypoint does not preserve signals
this_test_exit_status=0
run_test_process_and_send_signal "./bin/dojo --debug=true --test=true --image=alpine:3.15 -i=false sh -c \"echo 'will sleep' && sleep 1d\"" "will sleep"
test_value "${test_process_exit_status}" "2"
output_test_exit_status=$?
this_test_exit_status=$((this_test_exit_status+output_test_exit_status))
string_contains "${output}" "Caught signal: terminated"
# docker stop will take 10s, because sh does not preserves signals (from docker stop cli command to the docker container)
string_contains "${output}" "Stopping on signal"
string_contains "${output}" "docker stop testdojorunid"
string_contains "${output}" "Exit status from main work: 137"
string_contains "${output}" "Exit status from cleaning: 0"
string_contains "${output}" "Exit status from signals: 2"
output_test_exit_status=$?
this_test_exit_status=$((this_test_exit_status+output_test_exit_status))
echo "This test status: ${this_test_exit_status}"
rm -f "${test_file}"
log_test "----------------------------------------------------------"

if [[ "${this_test_exit_status}" != "${0}" ]]; then
  exit ${this_test_exit_status}
fi


log_test "Success"
