FROM golang:1.15.10

WORKDIR $GOPATH/src/github.com/weaveworks/weave-gitops
COPY . .
RUN go get -d -v ./...
RUN make all BINARY_NAME=wego

FROM alpine:3.5
WORKDIR /root/
COPY --from=0 $GOPATH/src/github.com/weaveworks/weave-gitops/ .
CMD ["wego"] 