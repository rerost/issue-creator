FROM golang:1.19-alpine

WORKDIR $GOPATH/src/github.com/rerost/issue-creator

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o issue-creator

COPY action.sh /action.sh

ENTRYPOINT ["/action.sh"]
