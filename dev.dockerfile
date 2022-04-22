FROM ubuntu
RUN apt-get update && apt-get install -yq ca-certificates
WORKDIR /app
ADD bin build
ENTRYPOINT /app/build/gitops
