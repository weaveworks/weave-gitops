FROM golang:1.15.10

WORKDIR $GOPATH/src/github.com/weaveworks/weave-gitops
COPY . .
RUN go get -d -v ./...
RUN make all BINARY_NAME=wego
CMD ["wego"] 