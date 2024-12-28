# Dojo image scripts

The `src` directory provides scripts which turn a usual docker image into a dojo docker image.

## Usage
Add the following directive in Dockerfile.

Please replace `TODO` with the latest Dojo version.

### Debian/Ubuntu
```
ENV DOJO_VERSION=TODO
RUN apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
  sudo git ca-certificates wget && \
  git clone --depth 1 -b ${DOJO_VERSION} https://github.com/kudulab/dojo.git /tmp/dojo_git && \
  /tmp/dojo_git/image_scripts/src/install.sh && \
  rm -r /tmp/dojo_git
```

### Alpine
```
ENV DOJO_VERSION=TODO
RUN apk add --no-cache bash shadow sudo git && \
  git clone --depth 1 -b ${DOJO_VERSION} https://github.com/kudulab/dojo.git /tmp/dojo_git && \
  /tmp/dojo_git/image_scripts/src/install.sh && \
  rm -r /tmp/dojo_git
```

Ignore the message : `Creating mailbox file: No such file or directory`.

### Additionally
You may want to add passwordless sudo for the dojo linux user:
```
echo 'dojo ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers
```

## What it does

The `src/install.sh` script:
   1. provides directories structure, used in dojo docker entrypoint script
   1. adds dojo linux user and group
   1. installs the dojo docker entrypoint script
   1. installs the script responsible for ensuring that dojo linux user has the same uid:gid as the `/dojo/work` volume owner


## Development
Build and test Dojo docker images:
```
cd image_scripts
./tasks build
./tasks test_scripts
./tasks e2e
```

Tests run by `./tasks test_scripts` do not use the images's entrypoints directly, but instead they test
 each script separately.

Tests run by `./tasks e2e` use the images' entrypoints directly. They test end user cases. They verify that
 `dojo <some-command>` is invocable and returns valid exit status and correct  output.
