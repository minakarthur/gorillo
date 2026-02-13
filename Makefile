.PHONY: run build tidy

run:
	go run ./cmd/gorillo

build:
	go build -o bin/gorillo ./cmd/gorillo

tidy:
	go mod tidy
