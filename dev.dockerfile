FROM ubuntu:24.04@sha256:59a458b76b4e8896031cd559576eac7d6cb53a69b38ba819fb26518536368d86
RUN apt-get update && apt-get install -yq ca-certificates
WORKDIR /app
ADD bin build
ENTRYPOINT /app/build/gitops
