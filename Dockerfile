FROM golang:1.19

WORKDIR $GOPATH/src/github.com/rerost/issue-creator

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install

COPY action.sh /action.sh

ENTRYPOINT ["/action.sh"]
