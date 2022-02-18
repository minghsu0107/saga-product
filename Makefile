# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install

all: build test

test: runtest
build: dep
	$(GOBUILD) -o server -v ./cmd/main.go
build-linux: dep
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o server -v ./cmd/main.go

runtest:
	$(GOTEST) -gcflags=-l -v -cover -coverpkg=./... -coverprofile=cover.out ./...
dep: wire
	$(shell $(GOCMD) env GOPATH)/bin/wire ./dep

wire:
	GO111MODULE=on $(GOINSTALL) github.com/google/wire/cmd/wire@v0.4.0

clean:
	$(GOCLEAN)
	rm -f server