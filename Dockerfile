FROM golang:1.15

WORKDIR $GOPATH/src/github.com/weaveworks/weave-gitops
COPY . .
RUN go get -d -v ./...
RUN go build -o wego
CMD ["wego"] 