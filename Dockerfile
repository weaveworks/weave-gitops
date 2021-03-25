FROM golang:1.15

WORKDIR $GOPATH/src/github.com/weaveworks/weave-gitops
COPY . .
RUN go get -d -v ./...
RUN make all
CMD ["wego"] 