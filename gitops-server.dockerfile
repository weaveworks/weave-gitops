# UI build
FROM node:22-bookworm@sha256:99981c3d1aac0d98cd9f03f74b92dddf30f30ffb0b34e6df8bd96283f62f12c6 AS ui
RUN apt-get update -y && apt-get install -y build-essential
RUN mkdir -p /home/app && chown -R node:node /home/app
WORKDIR /home/app
USER node
COPY --chown=node:node package*.json /home/app/
COPY --chown=node:node yarn.lock /home/app/
COPY --chown=node:node Makefile /home/app/
COPY --chown=node:node tsconfig.json /home/app/
COPY --chown=node:node .parcelrc /home/app/
COPY --chown=node:node .npmrc /home/app/
COPY --chown=node:node .yarn /home/app/.yarn
COPY --chown=node:node .yarnrc.yml /home/app/
RUN make node_modules
COPY --chown=node:node ui /home/app/ui
RUN --mount=type=cache,target=/home/app/ui/.parcel-cache make ui

# Go build
FROM golang:1.23.4@sha256:9820aca42262f58451f006de3213055974b36f24b31508c1baa73c967fcecb99 AS go-build

# Add known_hosts entries for GitHub and GitLab
RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN ssh-keyscan gitlab.com >> ~/.ssh/known_hosts

COPY Makefile /app/
WORKDIR /app
RUN go env -w GOCACHE=/go-cache
RUN --mount=type=cache,target=/gomod-cache \
    go env -w GOMODCACHE=/gomod-cache
COPY go.* /app/
RUN go mod download
COPY core /app/core
COPY pkg /app/pkg
COPY cmd /app/cmd
COPY api /app/api

# These are ARGS are defined here to minimise cache misses
# (cf. https://docs.docker.com/engine/reference/builder/#impact-on-build-caching)
# Pass these flags so we don't have to copy .git/ for those commands to work
ARG GIT_COMMIT="_unset_"
ARG LDFLAGS="-X localbuild=true"

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    LDFLAGS=${LDFLAGS##-X localbuild=true} GIT_COMMIT=$GIT_COMMIT make gitops-server

#  Distroless
FROM gcr.io/distroless/base@sha256:74ddbf52d93fafbdd21b399271b0b4aac1babf8fa98cab59e5692e01169a1348 AS runtime
COPY --from=ui /home/app/bin/dist/ /dist/
COPY --from=go-build /app/bin/gitops-server /gitops-server
COPY --from=go-build /root/.ssh/known_hosts /root/.ssh/known_hosts

ENTRYPOINT ["/gitops-server"]
