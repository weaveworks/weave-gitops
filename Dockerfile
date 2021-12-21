# UI build
FROM node:14-buster AS ui
RUN apt-get update -y && apt-get install -y build-essential
RUN mkdir -p /home/app && chown -R node:node /home/app
WORKDIR /home/app
USER node
COPY --chown=node:node package*.json /home/app/
COPY --chown=node:node Makefile /home/app/
RUN make node_modules
COPY --chown=node:node . /home/app/
RUN make ui

# Go build
FROM golang:1.17 AS go-build
COPY . /app
WORKDIR /app
# Remove flux to ensure the correct version is installed
RUN make clean
COPY --from=ui /home/app/cmd/gitops/ui/run/dist/ /app/cmd/gitops/ui/run/dist/
RUN make dependencies && make bin
# Add known_hosts entries for GitHub and GitLab
RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN ssh-keyscan gitlab.com >> ~/.ssh/known_hosts
# Add a kubectl
RUN apt-get install -y apt-transport-https ca-certificates curl openssh-client && \
    curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg \
    https://packages.cloud.google.com/apt/doc/apt-key.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] \
    https://apt.kubernetes.io/ kubernetes-xenial main" | tee /etc/apt/sources.list.d/kubernetes.list && \
    apt-get update && \
    apt-get install -y kubectl

# Distroless
FROM gcr.io/distroless/base
COPY --from=go-build /app/bin/gitops /gitops
COPY --from=go-build /root/.ssh/known_hosts /root/.ssh/known_hosts
COPY --from=go-build /usr/bin/kubectl /usr/bin/kubectl

ENTRYPOINT ["/gitops"]
