BUILDPATH=$(CURDIR)
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
API_NAME=BACKEND
LOWER_API_NAME=$(shell echo $(API_NAME) | tr A-Z a-z)

dir:
	@echo "full path: " $(BUILDPATH)

#.PHONY: build

build:
	@echo "Creating binary..."
	@CGO_ENABLED=1 go build -ldflags '-s -w' -o $(BUILDPATH)/build/bin/main main.go
	@echo " binary generated in build/bin/main"

test:
	@echo "Running tests..."
	@go test ./... -short ${PKG_LIST}
	@echo "OK"

test_coverage:
	@echo "Generating html coverage file..."
	@go test ./... -short -coverprofile=coverage.out ${PKG_LIST} 
	@echo "OK"


tidy:
	@go mod tidy


vendor:
	@go mod vendor

run:
	@go run main.go