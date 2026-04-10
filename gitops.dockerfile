ARG FLUX_VERSION=2.4.0
ARG FLUX_CLI=ghcr.io/fluxcd/flux-cli:v$FLUX_VERSION

# Alias for flux
FROM $FLUX_CLI@sha256:a9cb966cddc1a0c56dc0d57dda485d9477dd397f8b45f222717b24663471fd1f AS flux

# Go build
FROM golang:1.25.3@sha256:8c945d3e25320e771326dafc6fb72ecae5f87b0f29328cbbd87c4dff506c9135 AS go-build

# Add known_hosts entries for GitHub and GitLab
RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN ssh-keyscan gitlab.com >> ~/.ssh/known_hosts

COPY Makefile /app/
COPY tools /app/tools
WORKDIR /app
RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache
COPY go.* /app/
RUN --mount=type=cache,target=/gomod-cache \
    go mod download
COPY . /app

# These are ARGS are defined here to minimise cache misses
# (cf. https://docs.docker.com/engine/reference/builder/#impact-on-build-caching)
# Pass these flags so we don't have to copy .git/ for those commands to work
ARG LDFLAGS="-X localbuild=true"
ARG GIT_COMMIT="_unset_"

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    LDFLAGS=$LDFLAGS GIT_COMMIT=$GIT_COMMIT make gitops

# Distroless
FROM gcr.io/distroless/base@sha256:c83f022002fc917a92501a8c30c605efdad3010157ba2c8998a2cbf213299201 AS runtime
COPY --from=flux /usr/local/bin/flux /usr/local/bin/flux
COPY --from=go-build /app/bin/gitops /gitops
COPY --from=go-build /root/.ssh/known_hosts /root/.ssh/known_hosts

ENTRYPOINT ["/gitops"]
