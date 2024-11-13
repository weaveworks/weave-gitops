ARG FLUX_VERSION=0.24.1
ARG FLUX_CLI=ghcr.io/fluxcd/flux-cli:v$FLUX_VERSION

# Alias for flux
FROM $FLUX_CLI as flux

# Go build
FROM golang:1.23 AS go-build

# Add known_hosts entries for GitHub and GitLab
RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN ssh-keyscan gitlab.com >> ~/.ssh/known_hosts

COPY Makefile /app/
COPY tools /app/tools
WORKDIR /app
COPY go.* /app/
RUN go mod download
COPY . /app

# These are ARGS are defined here to minimise cache misses
# (cf. https://docs.docker.com/engine/reference/builder/#impact-on-build-caching)
# Pass these flags so we don't have to copy .git/ for those commands to work
ARG LDFLAGS="-X localbuild=true"
ARG GIT_COMMIT="_unset_"

RUN LDFLAGS=$LDFLAGS GIT_COMMIT=$GIT_COMMIT make gitops

# Distroless
FROM gcr.io/distroless/base as runtime
COPY --from=flux /usr/local/bin/flux /usr/local/bin/flux
COPY --from=go-build /app/bin/gitops /gitops
COPY --from=go-build /root/.ssh/known_hosts /root/.ssh/known_hosts

ENTRYPOINT ["/gitops"]
