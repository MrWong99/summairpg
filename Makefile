
all: build test

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o summairpg-linux cmd/summairpg/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o summairpg-darwin cmd/summairpg/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o summairpg.exe cmd/summairpg/main.go

test:
	go test ./...
