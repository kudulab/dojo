#!/bin/bash -e

dojouid=$(ls -n -d "${dojo_home}" | awk '{ print $3 }')

mkdir -p /run/user/${dojouid}
chown dojo:dojo /run/user/${dojouid}
