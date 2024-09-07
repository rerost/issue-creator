FROM golang:1.23-alpine

WORKDIR $GOPATH/src/github.com/rerost/issue-creator

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
COPY action.sh /action.sh
RUN go build -ldflags="-s -w" -trimpath -o /issue-creator .


ENTRYPOINT ["/action.sh"]
