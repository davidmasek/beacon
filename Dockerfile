FROM golang:1.23

WORKDIR /app

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=//root/go/pkg/mod

# pre-copy/cache go.mod for pre-downloading dependencies and only re-downloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/go/pkg/mod \
    go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build

ENTRYPOINT ["./beacon"]
CMD ["start"]
