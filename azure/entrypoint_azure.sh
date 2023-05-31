#!/bin/sh

##
# Created entrypoint for azure devops specially as it's only supports bash in posix mode:
# https://www.gnu.org/software/bash/manual/html_node/Bash-POSIX-Mode.html
# in our case if is using [ ] not [[ ]] also operation like ${WEAVE_REPO_BRANCH/#$REF_PREFIX} is not supported
##

# AzureDevops
export WEAVE_REPO_PROVIDER="azure-devops"
export WEAVE_REPO_URL="${SYSTEM_COLLECTIONURI}_git/${BUILD_REPOSITORY_NAME}"
export WEAVE_REPO_BRANCH="${BUILD_SOURCEBRANCHNAME}"
export WEAVE_REPO_SHA="${BUILD_SOURCEVERSION}"
export AZURE_PROJECT="${SYSTEM_TEAMPROJECT}"


exec weave-policy-validator "$@"
