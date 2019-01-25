# Dojo docker image

## Specification
A Dojo docker image is expected to:
   * have dojo linux user to run a short-running command as
   * have sudo and some shell installed (typically bash)
   * ensure the work directory is owned by the dojo user (run startup scripts fix-uid-gid script)
   * provide such an entrypoint that will
      * change current user to dojo user
      * run startup scripts
      * allow running interactively or not

Anyone can build and test a Dojo docker image in a simple way.

## Why not just use the official Docker images?
By using Dojo and Dojo Docker images:
   * Dojo takes care that your **current directory** (working directory) is mounted onto a container and that it is **owned by the same uid:gid in container as on host**. Using official Docker images, you usually run as root.
   * Dojo Docker images usually perform some **configuration on container start**. They may check that all the configuration files are setup in IdentityDirOuter, that necessary secrets exist or they may create new configuration files.

## How to deliver secrets into Dojo Docker images?
If secrets are **known** before a container is created:
   * they can be built-in into a Dojo Docker image - this is not recommended, but possible and easy
   * they can be set as environment variables. Dojo will preserve them into a container, no action from user is needed (just do not blacklist those variables)
   * they can be saved into files in IdentityDirOuter directory and the Dojo Docker image must take care to take them from `/dojo/identity` and deliver into expected destination
   * they can be saved into files and mounted into a container by setting `"DOJO_DOCKER_OPTIONS=-v /path/to/local/secret:/path/in/docker"`


If secrets are **not known** before a container is created:
   * a Dojo Docker image can have [Vault](https://www.vaultproject.io/) installed and on container start, secrets should be taken from Vault and saved to files. This is a nice idea, because after the container is deleted, the secrets files are deleted too (unless you mounted a volume under the secrets path)
   * use one Dojo Docker image to get secrets (and save them wherever you want to), then use another Dojo image to read the secrets and perform the main operation
