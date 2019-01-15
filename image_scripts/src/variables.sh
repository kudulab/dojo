#!/bin/bash

###########################################################################
# A file to keep any bash variables for Dojo to work.
# Override this file to change the below variables or add any new files in
# /etc/dojo.d/variables/ directory
###########################################################################

# dojo_work is the directory mounted as docker volume inside a docker container.
# From that directory we can infer uid and gid.
if [ -n "${DOJO_WORK_INNER}" ]; then
  export dojo_work="${DOJO_WORK_INNER}"
else
  export dojo_work="/dojo/work"
fi
export dojo_home="/home/dojo"
export dojo_identity="/dojo/identity"
export owner_username="dojo"
export owner_groupname="dojo"
