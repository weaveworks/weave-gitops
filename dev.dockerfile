FROM ubuntu
WORKDIR /app
ADD bin build
ADD server.rsa.crt build
ADD server.rsa.key build
ENTRYPOINT /app/build/gitops
