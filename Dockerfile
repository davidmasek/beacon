# Stage 1: Builder
# -----------------------------------
FROM golang:1.24 AS builder

WORKDIR /src

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/go/pkg/mod

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/go/pkg/mod \
    go mod download && go mod verify

COPY . .
ARG GIT_REF
ARG GIT_SHA
RUN echo "Specified version ${GIT_REF} ${GIT_SHA}"
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/beacon \
    -ldflags "-X 'github.com/davidmasek/beacon/conf.GitRef=${GIT_REF}' -X 'github.com/davidmasek/beacon/conf.GitSha=${GIT_SHA}'"

# Stage 2: Main
# -----------------------------------
FROM debian:bookworm-slim AS main

# Install:
# - certificates - needed to send emails securely
# - curl - useful for testing
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/beacon /app/beacon

WORKDIR /app

ENV BEACON_DB=/app/db/beacon.db

ENTRYPOINT ["/app/beacon"]
CMD ["start", "--config", "/app/beacon.yaml"]
