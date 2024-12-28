#!/bin/bash

test_file="dojo-signal-test-output"

function log_test() {
  echo "Signal test: $1"
}

function run_test_process_and_send_signal() {
    local test_process=${1?test_process not set}
    local kill_after_this_string=${2?kill_after_this_string not set}
    log_test "Will run test process: ${test_process}"
    log_test "Logging into this file: ${test_file}"

    # Start job in the background writing to a new file
    test_process_exit_status=0
    output=""
    ${test_process} >"${test_file}" 2>"${test_file}" &
    local pid=$!
    # local pgid=$(ps -o pgid ${pid} | tail -1 | tr -d " ")
    log_test "Initiated test process with PID: ${pid}, PGID: ${pgid}"
    local iteration_counter=0

    while true; do
      # the success
       if [[ $(grep "command will be" -c ${test_file}) == 1 ]] && [[ $(grep "${kill_after_this_string}" -c ${test_file}) == 4 ]]; then
            log_test "Dojo created the container and the container printed out the expected string. So we can kill the process now"
            break;
       fi

      test_process_output=$(cat "${test_file}")
      # by default `ps -p` works on ubuntu, but not on alpine;
      # on alpine you have to install this package: apk add --no-cache procps;
      # this was done in kudulab/inception-dojo:alpine-0.6.1 docker image
      if ps -p $pid > /dev/null; then
        # the need-to-wait-more case
        echo "The test process with $pid is running, so let's wait 1 s now, so that the container can produce the string: \"${kill_after_this_string}\". Then, we can kill the test process."
        sleep 1
      else
        # the failure
        log_test "Error while running the test process. The test process with $pid is no longer running:"
        log_test "Output from ${test_file}: ${test_process_output}"
        exit 1;
      fi
      if [ "${iteration_counter}" == "30" ]; then
        log_test "This takes suspiciously long. Let's read the ${test_file}"
        log_test "Output from ${test_file}: ${test_process_output}"
        # TODO: maybe just break here and cleanup;
      fi
      ((iteration_counter++))
    done

    # Kill the dojo process. If we kill the whole process group, no cleanup is done.
    ( set -x; kill -15 ${pid} && echo "successfully killed" || echo "failed to kill"; )
    log_test "Waiting for the test process to be finished"
    wait ${pid}
    test_process_exit_status=$?
    log_test "The test process is finished"

    log_test "----------------------------------------------------------"
    output=$(cat "${test_file}")
    log_test "Output from ${test_file}: ${output}"
    log_test "----------------------------------------------------------"
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
# Sometimes it may happen, that the docker service/daemon is not yet running.
# This especially happens when you just created this container, e.g.
# when running dojo in dojo (like in this test).
function wait_for_the_docker_daemon_to_be_running() {
  while (! docker ps > /dev/null ); do
    # Docker takes a few seconds to initialize
    echo "Waiting for the Docker daemon to launch..."
    sleep 1
  done
  echo "Docker daemon is running now"
}

# Test 1.: driver=docker, container's entrypoint does not preserve signals

## Setup
if [[ "${test_file}" == "" ]]; then
  log_test "test_file variable was unset; this happens on Alpine when running 'readlink -f' or 'realpath'"
  exit 1
fi
set -x; rm -f "${test_file}"; touch "${test_file}"; set +x;

wait_for_the_docker_daemon_to_be_running
this_test_exit_status=0
run_test_process_and_send_signal "./bin/dojo --debug=true --test=true --image=alpine:3.21 -i=false sh -c \"echo 'will sleep' && sleep 1d\"" "will sleep"

## Test checks
log_test "----------------------------------------------------------"
log_test "Now let's run some checks basing on the output from the file" >&2
test_value "${test_process_exit_status}" "2"
output_test_exit_status=$?
this_test_exit_status=$((this_test_exit_status+output_test_exit_status))
string_contains "${output}" "Caught signal 1: terminated"
# docker stop will take 10s, because sh does not preserves signals (from docker stop cli command to the docker container)
string_contains "${output}" "Stopping on signal"
string_contains "${output}" "docker stop testdojorunid"
string_contains "${output}" "Exit status from main work: 137"
string_contains "${output}" "Exit status from cleaning: 0"
string_contains "${output}" "Exit status from signals: 2"
output_test_exit_status=$?
this_test_exit_status=$((this_test_exit_status+output_test_exit_status))
echo "This test status: ${this_test_exit_status}"
# Let's not remove that file, so that we can look at it later for
# troubleshooting if needed:
# rm -f "${test_file}"
log_test "----------------------------------------------------------"

if [[ "${this_test_exit_status}" != "0" ]]; then
  log_test "Failure"
else
  log_test "Success"
fi

exit ${this_test_exit_status}
