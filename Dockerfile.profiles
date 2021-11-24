# Go build
FROM golang:1.17.1 AS builder
COPY . /app
WORKDIR /app
RUN make dependencies && make bin

# Distroless
FROM gcr.io/distroless/base
COPY --from=builder /app/bin/profiles-server /profiles-server
ENTRYPOINT ["/profiles-server"]
