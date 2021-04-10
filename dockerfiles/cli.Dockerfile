# syntax=docker/dockerfile:experimental
FROM golang:latest as build

WORKDIR /go/src/app

# prefetch go.mod dependencies and store them in cache
COPY ./go.mod .
COPY ./go.sum .
RUN --mount=type=cache,target=/go/pkg/mod go mod download -x

# build my-app executable and store build files in cache
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    go build -o ./bin/facebook-liker ./cmd/cli

# copy executable to new container, so final container size will be less and source code won't be exposed
# through docker history command
WORKDIR /go/src/app

FROM alpine:latest
COPY --from=build /go/src/app/data/users.csv /go/src/app/data/users.csv
COPY --from=build /go/src/app/bin /go/src/app/bin

CMD ["/go/src/app/bin/facebook-liker", "-a", "selenium-hub:4444"]