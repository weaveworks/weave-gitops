FROM ubuntu
WORKDIR /app
ADD bin build
ADD tls /app
ENTRYPOINT /app/build/gitops
