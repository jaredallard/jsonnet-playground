# syntax=docker/dockerfile:1.0-experimental
ARG GO_VERSION
FROM golang:1.18.3-alpine AS build
ARG VERSION
ENV GOCACHE "/go-build-cache"
ENV NODECACHE "/node-build-cache"
ENV CGO_ENABLED 0
WORKDIR /src

# Need git for VCS information in the binary, ca-certs for the scratch image
RUN apk --update add git ca-certificates nodejs yarn

# Copy our source code into the container for building
COPY . .

# Cache dependencies across builds
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg go mod download

# Build our application, caching the go build cache, but also using
# the dependency cache from earlier.
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg --mount=type=cache,target=${GOCACHE} \
  mkdir -p bin; go build -ldflags "-w -s" -o /src/bin/ -v ./cmd/...

# Build the UI
WORKDIR /src/web
RUN --mount=type=cache,target=${NODECACHE} yarn config set cache-folder ${NODECACHE}
RUN --mount=type=cache,target=${NODECACHE} --mount=type=cache,target=/src/node_moules yarn --frozen-lockfile && \
  yarn build && yarn run next export

FROM busybox
ENTRYPOINT [ "/usr/local/bin/jsonnet-playground" ]
WORKDIR /app
ENV ZONEINFO=/zoneinfo.zip

# Dependencies of the binary
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /src/web/out /app/web/out

# Binary
COPY --from=build ./src/bin/jsonnet-playground /usr/local/bin/jsonnet-playground
