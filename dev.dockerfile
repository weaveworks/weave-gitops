FROM ubuntu:24.04@sha256:728785b59223d755e3e5c5af178fab1be7031f3522c5ccd7a0b32b80d8248123
RUN apt-get update && apt-get install -yq ca-certificates
WORKDIR /app
ADD bin build
ENTRYPOINT /app/build/gitops
