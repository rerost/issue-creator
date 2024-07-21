FROM golang:1.12-alpine AS builder

WORKDIR $GOPATH/src/github.com/rerost/issue-creator

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o /issue-creator

FROM alpine:3.18.3

COPY action.sh /action.sh
COPY --from=builder /issue-creator /issue-creator

ENTRYPOINT ["/action.sh"]
