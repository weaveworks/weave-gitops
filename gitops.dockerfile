ARG FLUX_VERSION=2.4.0
ARG FLUX_CLI=ghcr.io/fluxcd/flux-cli:v$FLUX_VERSION

# Alias for flux
FROM $FLUX_CLI@sha256:a9cb966cddc1a0c56dc0d57dda485d9477dd397f8b45f222717b24663471fd1f AS flux

# Go build
FROM golang:1.23@sha256:574185e5c6b9d09873f455a7c205ea0514bfd99738c5dc7750196403a44ed4b7 AS go-build

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
FROM gcr.io/distroless/base@sha256:e9d0321de8927f69ce20e39bfc061343cce395996dfc1f0db6540e5145bc63a5 AS runtime
COPY --from=flux /usr/local/bin/flux /usr/local/bin/flux
COPY --from=go-build /app/bin/gitops /gitops
COPY --from=go-build /root/.ssh/known_hosts /root/.ssh/known_hosts

ENTRYPOINT ["/gitops"]
