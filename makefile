.PHONY: run build clean frontend

run:
	go run cmd/server/main.go

frontend:
	cd frontend && npm run dev

build:
	go build -o server cmd/server/main.go

clean:
	rm -f server

lint:
	golangci-lint run ./...
	cd frontend && npm run lint
