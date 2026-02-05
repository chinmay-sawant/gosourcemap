.PHONY: run build test clean tidy

run:
	go run cmd/server/main.go

build:
	go build -o bin/order-service cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

tidy:
	go mod tidy
