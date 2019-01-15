#!/bin/bash

set -ue

###########################################################################
# This file ensures that files are mapped from dojo_identity
# into dojo_home. Fails if any required secret or configuration file is missing.
###########################################################################

# 1. ssh directory, copy it with all the secrets, particularly id_rsa
if [ ! -d "${dojo_identity}/.ssh" ]; then
  echo "WARN: ${dojo_identity}/.ssh does not exist"
fi
if [ ! -f "${dojo_identity}/.ssh/id_rsa" ]; then
  echo "WARN: ${dojo_identity}/.ssh/id_rsa does not exist"
fi
cp -ar "${dojo_identity}/.ssh" "${dojo_home}"

# 2. ssh config, create it

echo "StrictHostKeyChecking no
UserKnownHostsFile /dev/null
" > "${dojo_home}/.ssh/config"
