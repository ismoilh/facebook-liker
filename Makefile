all: build run

build:
	go build -o ./bin/facebook-liker ./cmd/server

run:
	./bin/facebook-liker -a 0.0.0.0:7070 -s 127.0.0.1:4444

docker-build-run:
	docker compose up --build --force-recreate