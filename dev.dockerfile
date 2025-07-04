FROM ubuntu:24.04@sha256:440dcf6a5640b2ae5c77724e68787a906afb8ddee98bf86db94eea8528c2c076
RUN apt-get update && apt-get install -yq ca-certificates
WORKDIR /app
ADD bin build
ENTRYPOINT /app/build/gitops
