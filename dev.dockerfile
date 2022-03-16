FROM ubuntu
WORKDIR /app
ADD bin build
ADD localhost.pem build
ADD localhost-key.pem build
ENTRYPOINT /app/build/gitops
