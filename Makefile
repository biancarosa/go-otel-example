.PHONY: build run clean docker-build docker-up docker-down docker-logs

build:
	go mod tidy
	go build -o api

run: build
	./api

clean:
	rm -f api

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

test:
	go test ./...

load:
	./scripts/load.sh