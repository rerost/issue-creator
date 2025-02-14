FROM golang:1.24-alpine AS builder

WORKDIR $GOPATH/src/github.com/rerost/issue-creator

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    go build -ldflags="-s -w" -trimpath -o /issue-creator .

FROM alpine:3.21.3

COPY action.sh /action.sh
COPY --from=builder /issue-creator /issue-creator

ENTRYPOINT ["/action.sh"]
