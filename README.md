# issue-creator
[![codecov](https://codecov.io/gh/rerost/issue-creator/branch/master/graph/badge.svg?token=fh77gKdsoh)](https://codecov.io/gh/rerost/issue-creator)

Automated GitHub Issue Scheduler — CLI & GitHub Action for recurring tasks

e.g.
- Template Issue: https://github.com/rerost/issue-creator/issues/1
- Template Manifest: WIP

## Breaking Change
v0.1.x -> v0.2.x: The behavior has changed from moving to the Archive Category when closing a GitHub Discussion to simply closing the Discussion

## Install
```
$ go install github.com/rerost/issue-creator@latest
$ issue-creator
```

## Usage
Run by GitHub Actions

Example
```yaml
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

```yaml
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
