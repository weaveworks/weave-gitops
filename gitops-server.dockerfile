# UI build
FROM node:22-bookworm@sha256:c7fd844945a76eeaa83cb372e4d289b4a30b478a1c80e16c685b62c54156285b AS ui
RUN apt-get update -y && apt-get install -y build-essential python3 g++
RUN npm install -g node-gyp
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
RUN make ui

# Go build
FROM golang:1.24.1@sha256:52ff1b35ff8de185bf9fd26c70077190cd0bed1e9f16a2d498ce907e5c421268 AS go-build

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
FROM gcr.io/distroless/base@sha256:125eb09bbd8e818da4f9eac0dfc373892ca75bec4630aa642d315ecf35c1afb7 AS runtime
COPY --from=ui /home/app/bin/dist/ /dist/
COPY --from=go-build /app/bin/gitops-server /gitops-server
COPY --from=go-build /root/.ssh/known_hosts /root/.ssh/known_hosts

ENTRYPOINT ["/gitops-server"]
