# syntax=docker/dockerfile:experimental
FROM golang:latest as build

WORKDIR /go/src/app

# prefetch go.mod dependencies and store them in cache

COPY ./go.mod .
COPY ./go.sum .

# build my-app executable and store build files in cache
COPY . .
RUN CGO_ENABLED=0 \
    go build -o ./bin/facebook-liker ./cmd/server

# copy executable to new container, so final container size will be less and source code won't be exposed
# through docker history command
FROM alpine:latest
COPY --from=build /go/src/app/data/users.csv /go/src/app/data/users.csv
COPY --from=build /go/src/app/bin /go/src/app/bin

EXPOSE 7070

CMD ["/go/src/app/bin/facebook-liker", "-a", "0.0.0.0:7070", "-s", "selenium-hub:4444"]