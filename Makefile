.PHONY: build test clean run-server run-cli

build:
	go build -o bin/pi_test ./cmd/pi/ai-agent-session/main.go

test:
	go test ./...

clean:
	rm -rf bin/

run-cli:
	go run ./cmd/pi/ai-agent-session/main.go
