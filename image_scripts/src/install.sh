#!/bin/bash

# This script helps to make an dojo docker image.
# It installs:
# * dojo scripts into /etd/dojo.d/scripts/
# * dojo variables into /etd/dojo.d/variables/
# * dojo entrypoint into /usr/bin/entrypoint.sh
# * adds dojo linux user and group
# Works on Ubuntu/Debian/Alpine.

# Absolute path to this script
script_path=$(readlink -f "$BASH_SOURCE")
# Absolute path to a directory this script is in
script_dir=$(dirname "${script_path}")

cp "${script_dir}/entrypoint.sh" /usr/bin/entrypoint.sh
mkdir -p /etc/dojo.d
mkdir -p /etc/dojo.d/scripts
mkdir -p /etc/dojo.d/variables
# 50 is because user may want to do things before and after home and work
# directories ownership was fixed. Also user may wish to delete/replace this script.
cp "${script_dir}/50-fix-uid-gid.sh" /etc/dojo.d/scripts/50-fix-uid-gid.sh
cp "${script_dir}/variables.sh" /etc/dojo.d/variables/50-variables.sh

# Add dojo user and group
groupadd --gid 1000 dojo
useradd --home-dir /home/dojo --uid 1000 --gid 1000 --shell /bin/bash dojo
usermod -a -G dojo dojo # adds user: dojo to group: dojo

mkdir -p /home/dojo
chown dojo:dojo /home/dojo
