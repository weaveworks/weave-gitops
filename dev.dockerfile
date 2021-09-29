FROM golang:1.16-buster

RUN apt-get update -y && apt-get install -y build-essential

WORKDIR $GOPATH/src/github.com/weaveworks/weave-gitops

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go mod tidy

CMD ["sh","-c","go run cmd/gitops/main.go ui run"]

EXPOSE 9001