FROM ubuntu
WORKDIR /app
ADD bin build
ENTRYPOINT /app/build/gitops
