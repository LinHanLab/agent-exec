all: compile quality

compile:
	go build -o app ./cmd

quality:
	go test ./...
	go fmt ./...
	golangci-lint run
