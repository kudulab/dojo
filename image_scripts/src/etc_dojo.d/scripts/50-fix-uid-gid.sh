#!/bin/bash

###########################################################################
# This file ensures that dojo user has the same uid and gid as /dojo/work directory.
# Used as fix-uid-gid solution in docker, almost copied from:
# https://github.com/tomzo/docker-uid-gid-fix/blob/master/fix-uid-gid.sh
###########################################################################

function dojo_50fixuidgid_log_info {
  if [[ "${DOJO_LOG_LEVEL}" != "silent" ]] && [[ "${DOJO_LOG_LEVEL}" != "error" ]] && [[ "${DOJO_LOG_LEVEL}" != "warn" ]]; then
    echo -e "$(date "+%d-%m-%Y %T") Dojo 50-fix-uid-gid info: ${1}" >&2
  fi
}
function dojo_50fixuidgid_log_error {
  if [[ "${DOJO_LOG_LEVEL}" != "silent" ]] ; then
    echo -e "$(date "+%d-%m-%Y %T") Dojo 50-fix-uid-gid error: ${1}" >&2
  fi
}

if [[ -z "${dojo_work}" ]]; then
    dojo_50fixuidgid_log_error "dojo_work not specified"
    exit 1;
fi
if [[ ! -d "${dojo_work}" ]]; then
    dojo_50fixuidgid_log_error "$dojo_work does not exist, expected to be mounted as docker volume"
    exit 1;
fi

if [[ -z "${dojo_home}" ]]; then
    dojo_50fixuidgid_log_error "dojo_home not set"
    exit 1;
fi

if [[ -z "${owner_username}" ]]; then
    dojo_50fixuidgid_log_error "Username not specified"
    exit 1;
fi
if [[ -z "${owner_groupname}" ]]; then
    dojo_50fixuidgid_log_error "Groupname not specified"
    exit 1;
fi

if ! getent passwd "${owner_username}" >/dev/null 2>&1; then
    dojo_50fixuidgid_log_error "User ${owner_username} does not exist"
    exit 1;
fi

if ! getent passwd "${owner_groupname}" >/dev/null 2>&1; then
    dojo_50fixuidgid_log_error "Group ${owner_groupname} does not exist"
    exit 1;
fi



# use -n option which is the same as --numeric-uid-gid on Debian/Ubuntu,
# but on Alpine, there is no --numeric-uid-gid option
# Why we are sudo-ing as dojo? To deal with OSX legacy volume driver; see more at https://github.com/kudulab/dojo/issues/8
newuid=$(sudo -u "${owner_username}" ls -n -d "${dojo_work}" | awk '{ print $3 }')
newgid=$(sudo -u "${owner_username}" ls -n -d "${dojo_work}" | awk '{ print $4 }')
olduid=$(ls -n -d ${dojo_home} | awk '{ print $3 }')
oldgid=$(ls -n -d ${dojo_home} | awk '{ print $4 }')

if [[ "${olduid}" == "${newuid}" ]] && [[ "${oldgid}" == "${newgid}" ]]; then
    dojo_50fixuidgid_log_info "olduid == newuid == ${newuid}, nothing to do"
elif [[ "0" == "${newuid}" ]] && [[ "0" == "${newgid}" ]]; then
    # We are on gRPC FUSE driver on Mac - do nothing
    # Or dojo was executed from host where current directory is owned by the root - not supported use case
    dojo_50fixuidgid_log_info "Assuming docker running with gRPC FUSE driver on Mac"
else
    if [[ "${DOJO_LOG_LEVEL}" != "silent" ]] && [[ "${DOJO_LOG_LEVEL}" != "error" ]] && [[ "${DOJO_LOG_LEVEL}" != "warn" ]]; then
      set -x
    fi
    ( usermod -u "${newuid}" "${owner_username}"; groupmod -g "${newgid}" "${owner_groupname}"; )
    ( chown ${newuid}:${newgid} -R "${dojo_home}"; )
    if [[ "${DOJO_LOG_LEVEL}" != "silent" ]] && [[ "${DOJO_LOG_LEVEL}" != "error" ]] && [[ "${DOJO_LOG_LEVEL}" != "warn" ]]; then
      set +x
    fi
fi

# do not chown the "$dojo_work" directory, it already has proper uid and gid,
# besides, when "$dojo_work" is very big, chown would take long time
