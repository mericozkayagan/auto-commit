# Go parameters

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=autocommit
BINARY_DIR=bin


.PHONY: all build test clean

all: build

build:
	@echo "Building $(BINARY_NAME)"
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME)
	cp $(BINARY_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

test:
	@echo "Running tests"
	$(GOTEST) -v ./...

clean:
	@echo "Cleaning"
	$(GOCLEAN)
	rm -f $(BINARY_DIR)/$(BINARY_NAME)
	rm -rf /usr/local/bin/$(BINARY_NAME)
