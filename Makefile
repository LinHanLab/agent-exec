all: compile quality

compile:
	go build -o app .

quality:
	go test ./...
	go fmt ./...
	golangci-lint run
