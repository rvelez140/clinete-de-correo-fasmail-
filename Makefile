.PHONY: build run dev up down logs test clean

build:
	go build -o bin/fasmail-panel ./cmd/server

run: build
	./bin/fasmail-panel

dev:
	FASMAIL_DEV=true go run ./cmd/server

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

test:
	go test ./... -v -race

clean:
	docker compose down -v
	rm -rf bin/
