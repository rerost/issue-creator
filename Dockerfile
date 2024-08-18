FROM golang:1.23-alpine AS builder

WORKDIR $GOPATH/src/github.com/rerost/issue-creator

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -trimpath -o /issue-creator .

FROM alpine:3.20.2

COPY action.sh /action.sh
COPY --from=builder /issue-creator /issue-creator

ENTRYPOINT ["/action.sh"]
