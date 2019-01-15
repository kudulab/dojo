#!/bin/bash

###########################################################################
# This file ensures that dojo user has the same uid and gid as /dojo/work directory.
# Used as fix-uid-gid solution in docker, almost copied from:
# https://github.com/tomzo/docker-uid-gid-fix/blob/master/fix-uid-gid.sh
###########################################################################

if [[ -z "${dojo_work}" ]]; then
    echo "dojo_work not specified"
    exit 1;
fi
if [[ ! -d "${dojo_work}" ]]; then
    echo "$dojo_work does not exist, expected to be mounted as docker volume"
    exit 1;
fi

if [[ -z "${dojo_home}" ]]; then
    echo "dojo_home not set"
    exit 1;
fi

if [[ -z "${owner_username}" ]]; then
    echo "Username not specified"
    exit 1;
fi
if [[ -z "${owner_groupname}" ]]; then
    echo "Groupname not specified"
    exit 1;
fi

if ! getent passwd "${owner_username}" >/dev/null 2>&1; then
    echo "User ${owner_username} does not exist"
    exit 1;
fi

if ! getent passwd "${owner_groupname}" >/dev/null 2>&1; then
    echo "Group ${owner_groupname} does not exist"
    exit 1;
fi

# use -n option which is the same as --numeric-uid-gid on Debian/Ubuntu,
# but on Alpine, there is no --numeric-uid-gid option
newuid=$(ls -n -d "${dojo_work}" | awk '{ print $3 }')
newgid=$(ls -n -d "${dojo_work}" | awk '{ print $4 }')

( set -x; usermod -u "${newuid}" "${owner_username}"; groupmod -g "${newgid}" "${owner_groupname}"; )
( set -x; chown ${newuid}:${newgid} -R "${dojo_home}"; )

# do not chown the "$dojo_work" directory, it already has proper uid and gid,
# besides, when "$dojo_work" is very big, chown would take much time
