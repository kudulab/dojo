#!/bin/bash

###########################################################################
# Dojo Docker container entrypoint
###########################################################################

set -e

bold='\033[1m'
green='\033[0;32m'
green_bold='\033[1;32m'
noformat='\033[0m'
function dojo_entrypoint_log_info {
  echo -e "${bold}$(date "+%d-%m-%Y %T") Dojo entrypoint info:${noformat} ${1}" >&2
}
function dojo_entrypoint_log_info_green {
  echo -e "${green_bold}$(date "+%d-%m-%Y %T") Dojo entrypoint info:${green} ${1}${noformat}" >&2
}

# source any additional scripts with environment variables
for SCRIPT in /etc/dojo.d/variables/* ; do
	if [ -f ${SCRIPT} ] ; then
    dojo_entrypoint_log_info "Sourcing: ${SCRIPT}"
		chmod +x ${SCRIPT}
		source ${SCRIPT}
	fi
done
# run any additional scripts to setup custom configuration files or secrets
# or source any files or wait for linux daemons
for SCRIPT in /etc/dojo.d/scripts/* ; do
	if [ -f ${SCRIPT} ] ; then
    dojo_entrypoint_log_info "Running: ${SCRIPT}"
		chmod +x ${SCRIPT}
		${SCRIPT}
	fi
done

# Note that any log messages go to stderr, so that we can save return value
# of dojo command into a bash variable. E.g.
# version=$(dojo some-command-to-get-version)
# This is not needed if you don't intend to save it.
if [ -t 0 ] ; then
    # interactive shell
    dojo_entrypoint_log_info_green "dojo init finished (interactive shell)" >&2

    # No "set -e" here, you don't want to be logged out when sth returns not 0
    # in interactive shell.
    set +e
else
    # not interactive shell
    dojo_entrypoint_log_info_green "dojo init finished (not interactive shell)" >&2
    set -e
fi

if [[ -n "${this_image_name}" ]] || [[ -n "${this_image_tag}" ]]; then
  # variables set
  dojo_entrypoint_log_info_green "using ${this_image_name}:${this_image_tag}" >&2
fi

sudo -E -H -u dojo /bin/bash -lc "$@"
