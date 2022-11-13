[![codecov](https://codecov.io/gh/MagalixTechnologies/weave-iac-validator/branch/main/graph/badge.svg?token=T2PlPCEuvG)](https://codecov.io/gh/MagalixTechnologies/weave-iac-validator)

# Weaveworks Infrastructure as Code Validator

Validates infrastucture as code against weave policies

## Supported Kustomizations
- [x] Helm
- [x] Kustomize

## Supported CI/CD
- [x] [Github](#github)
- [x] [Gitlab](#gitlab)
- [x] [Bitbucket](#bitbucket)
- [x] [Circle CI](#circle-ci)
- [x] [Azure Devops](#azure-devops)

## Usage
```bash
USAGE:
   app [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path value                       path to resources kustomization directory
   --helm-values-file value           path to resources helm values file
   --policies-path value              path to policies kustomization directory
   --policies-helm-values-file value  path to policies helm values file
   --git-repo-provider value          git repository provider [$WEAVE_REPO_PROVIDER]
   --git-repo-url value               git repository url [$WEAVE_REPO_URL]
   --git-repo-branch value            git repository branch [$WEAVE_REPO_BRANCH]
   --git-repo-sha value               git repository commit sha [$WEAVE_REPO_SHA]
   --git-repo-token value             git repository token [$WEAVE_REPO_TOKEN]
   --azure-project value              azure project name [$AZURE_PROJECT]
   --sast value                       save result as gitlab sast format
   --sarif value                      save result as sarif format
   --json value                       save result as json format
   --generate-git-report              generate git report if supported (default: false) [$WEAVE_GENERATE_GIT_PROVIDER_REPORT]
   --remediate                        auto remediate resources if possible (default: false)
   --no-exit-error                    exit with no error (default: false)
   --help, -h                         show help (default: false)
   --version, -v                      print the version (default: false)
```

## Examples

### Github
See how to setup the [Github Action](https://github.com/weaveworks/weave-action)

### Gitlab

```yaml
weave:
  image:
    name: weaveworks/weave-iac-validator:v1
  script:
  - weave-validator --path <path to resources> --policies-path <path to policies>
```

#### Enable Auto Remediation

```yaml
  script:
  - weave-validator --path <path to resources> --policies-path <path to policies> --git-repo-token $GITLAB_TOKEN --remediate
```

#### Enable Static Application Security Testing

```yaml
stages:
  - weave
  - sast

weave:
  stage: weave
  image:
    name: weaveworks/weave-iac-validator:v1
  script:
  - weave-validator <path to resources> --policies-path <path to policies> --sast sast.json
  artifacts:
    when: on_failure
    paths:
    - sast.json

upload_sast:
  stage: sast
  when: always
  script:
  - echo "creating sast report" 
  artifacts:
    reports:
      sast: sast.json
```


### Bitbucket

```yaml
pipelines:
  default:
    - step:
        name: 'Weaveworks'
        image: weaveworks/weave-iac-validator:v1
        script:
          - weave-validator --path <path to resources> --policies-path <path to policies>
```
#### Enable Auto Remediation

```yaml
  script:
    - weave-validator --path <path to resources> --policies-path <path to policies> --git-repo-token $TOKEN --remediate
```

#### Create Pipeline Report

```yaml
  script:
    - weave-validator --path <path to resources> --policies-path <path to policies> --git-repo-token $TOKEN -generate-git-report
```

### Circle CI

```yaml
jobs:
  weave:
    docker:
    - image: weaveworks/weave-iac-validator:v1
    steps:
    - checkout
    - run:
        command: weave-validator --path <path to resources> --policies-path <path to policies>
```

### Azure DevOps

```yaml
jobs:
  weave:
    docker:
    - image: weaveworks/weave-iac-validator:v1
    steps:
    - checkout
    - run:
        command: weave-validator --path <path to resources> --policies-path <path to policies>
```

#### Enable Auto Remediation

```yaml
    - run:
        command: weave-validator --path <path to resources> --policies-path <path to policies> --git-repo-token ${GITHUB_TOKEN} --remediate
```
