.PHONY: run build clean

run:
	go run cmd/server/main.go

build:
	go build -o server cmd/server/main.go

clean:
	rm -f server
