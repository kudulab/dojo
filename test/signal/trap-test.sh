#!/bin/bash

function on_sigint() {
    echo "Received SIGINT"
    exit 1
}
function on_exit() {
    echo "Received exit"
    exit 1
}

trap on_exit EXIT
trap on_sigint SIGINT

while true; do date; sleep 1; done