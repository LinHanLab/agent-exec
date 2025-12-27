all: compile quality

compile:
	go build -o agent-exec ./cmd/agent-exec

install:
	go install ./cmd/agent-exec

quality:
	go test ./...
	go fmt ./...
	golangci-lint run
