FROM golang:1.16-buster

RUN apt-get update -y && apt-get install -y build-essential

WORKDIR $GOPATH/src/github.com/weaveworks/weave-gitops

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s

COPY go.mod .
COPY go.sum .

COPY pkg ./pkg
COPY cmd ./cmd
COPY api ./api
COPY manifests ./manifests
COPY main.go .
COPY .air.toml .

RUN go mod download

CMD ["sh","-c","go run cmd/gitops/main.go ui run"]

EXPOSE 9001