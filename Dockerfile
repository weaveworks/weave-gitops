# UI build
FROM node:14-buster AS ui
RUN apt-get update -y && apt-get install -y build-essential
RUN mkdir -p /home/app/node_modules && chown -R node:node /home/app
WORKDIR /home/app
USER node
COPY --chown=node:node package*.json /home/app/
RUN npm install
COPY --chown=node:node . /home/app/
RUN make ui

# Go build
FROM golang:1.16 AS go-build
COPY . /app
COPY --from=ui /home/app/cmd/gitops/ui/run/dist/ /app/cmd/gitops/ui/run/dist/
WORKDIR /app
RUN make dependencies && make bin

# Distroless
FROM gcr.io/distroless/base
COPY --from=go-build /app/bin/gitops /gitops
ENTRYPOINT ["/gitops"]
