all: compile quality

compile:
	go build -o app ./cmd/agent-exec

install:
	go install ./cmd/agent-exec

quality:
	go test ./...
	go fmt ./...
	golangci-lint run
