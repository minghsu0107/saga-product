# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

all: test build

test: runtest
build: dep
	$(GOBUILD) -o server -v ./cmd/main.go
build-linux: dep
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o server -v ./cmd/main.go

runtest:
	$(GOTEST) -v ./...
dep: wire
	$(shell $(GOCMD) env GOPATH)/bin/wire ./dep

wire:
	GO111MODULE=on $(GOGET) -u github.com/google/wire/cmd/wire@v0.4.0

clean:
	$(GOCLEAN)
	rm -f server