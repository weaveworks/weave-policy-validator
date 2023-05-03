#!/bin/sh

REF_PREFIX="refs/heads/"

# Github
if [[ ${GITHUB_ACTIONS} ]]
then
    export WEAVE_REPO_PROVIDER="github"
    export WEAVE_REPO_URL="${GITHUB_REPOSITORY}"
    export WEAVE_REPO_BRANCH="${GITHUB_HEAD_REF:-$GITHUB_REF}"
    export WEAVE_REPO_SHA="${GITHUB_SHA}"
    export WEAVE_REPO_TOKEN="${GITHUB_TOKEN}"

# Gitlab
elif [[ ${GITLAB_CI} ]]
then
    export WEAVE_REPO_PROVIDER="gitlab"
    export WEAVE_REPO_URL="${CI_PROJECT_PATH}"
    export WEAVE_REPO_BRANCH="${CI_COMMIT_REF_NAME}"
    export WEAVE_REPO_SHA="${CI_COMMIT_SHA}"
    export WEAVE_REPO_TOKEN="${CI_JOB_TOKEN}"

# Bitbucket
elif [[ ${BITBUCKET_REPO_FULL_NAME} ]]
then
    export WEAVE_REPO_PROVIDER="bitbucket"
    export WEAVE_REPO_URL="${BITBUCKET_REPO_FULL_NAME}"
    export WEAVE_REPO_BRANCH="${BITBUCKET_BRANCH}"
    export WEAVE_REPO_SHA="${BITBUCKET_COMMIT}"

# CircleCI
elif [[ ${CIRCLECI} ]]
then
    export WEAVE_REPO_PROVIDER="github"
    export WEAVE_REPO_URL="${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}"
    export WEAVE_REPO_BRANCH="${CIRCLE_BRANCH}"
    export WEAVE_REPO_SHA="${CIRCLE_SHA1}"

fi

export WEAVE_REPO_BRANCH=${WEAVE_REPO_BRANCH/#$REF_PREFIX}

exec weave-policy-validator "$@"
