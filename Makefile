all: compile quality

compile:
	go build -o app .

install:
	go install .

quality:
	go test ./...
	go fmt ./...
	golangci-lint run
