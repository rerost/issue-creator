FROM golang:1.15

WORKDIR $GOPATH/src/github.com/rerost/issue-creator

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install
