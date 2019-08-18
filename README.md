# issue-creator
This is a CLI tool for automatically generating periodically created issues.

## Flow
![image](https://user-images.githubusercontent.com/5201588/63219703-a4848b00-c1b2-11e9-90a7-aa2a4920d47b.png)

## Install
```
$ GO111MODULE=on go get github.com/rerost/issue-creator
$ issue-creator
```

## Useage
NOTE: please set `GithubAccessToken` for create issue, `K8sCommands` for schedule issue

```
issue-creator render https://github.com/wantedly/issue-creator/issues/1
issue-creator create https://github.com/wantedly/issue-creator/issues/1
issue-creator schedule render '30 5 * * 1' https://github.com/wantedly/issue-creator/issues/1
issue-creator schedule apply '30 5 * * 1' https://github.com/wantedly/issue-creator/issues/1
```
