# issue-creator
[![codecov](https://codecov.io/gh/rerost/issue-creator/branch/master/graph/badge.svg?token=fh77gKdsoh)](https://codecov.io/gh/rerost/issue-creator)

This is a CLI tool for automatically generating periodically created issues.

e.g.
- Template Issue: https://github.com/rerost/issue-creator/issues/1
- Template Manifest: WIP

## Breaking Change
v0.1.x -> v0.2.x: The behavior has changed from moving to the Archive Category when closing a GitHub Discussion to simply closing the Discussion


## Flow
![image](https://user-images.githubusercontent.com/5201588/63219703-a4848b00-c1b2-11e9-90a7-aa2a4920d47b.png)

## Install
```
$ go install github.com/rerost/issue-creator@latest
$ issue-creator
```

## Usage
NOTE: please set `GithubAccessToken` for create issue, `K8sCommands` for schedule issue

```
issue-creator render https://github.com/rerost/issue-creator/issues/1
issue-creator create https://github.com/rerost/issue-creator/issues/1
issue-creator create https://github.com/rerost/issue-creator/issues/1 --CloseLastIssue
issue-creator schedule render '30 5 * * 1' https://github.com/rerost/issue-creator/issues/1
issue-creator schedule apply '30 5 * * 1' https://github.com/rerost/issue-creator/issues/1
issue-creator schedule apply '30 5 * * 1' https://github.com/rerost/issue-creator/issues/1 --CloseLastIssue
```

## Discussion
```
issue-creator create https://github.com/rerost/issue-creator/discussions/48
```

## Use from GitHub Actions
Example
```
on:
  schedule:
    - cron: "0 0 * * MON"
  workflow_dispatch: {}

jobs:
  create-issue:
    runs-on: ubuntu-latest
    steps:
      - uses: rerost/issue-creator@v0.4
        with:
          template-issue: 1 # https://github.com/rerost/issue-creator/issues/1
          discussion: true # Required if you want to create a discussion
```

or

```
on:
  schedule:
    - cron: "0 0 * * MON"
  workflow_dispatch: {}

jobs:
  create-issue:
    runs-on: ubuntu-latest
    steps:
      - uses: rerost/issue-creator@v0.4
        with:
          template-issue-url: https://github.com/rerost/issue-creator/issues/1
```

## Development
### Release
```bash
$ ./release.sh <TAG>

// e.g `./release.sh v0.3.1`
```

### Update token
https://github.com/rerost/issue-creator/issues/111
