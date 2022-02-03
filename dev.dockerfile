FROM ubuntu
RUN apt-get update && apt-get install -y ca-certificates openssl
WORKDIR /app
ADD bin build
ENTRYPOINT /app/build/gitops
