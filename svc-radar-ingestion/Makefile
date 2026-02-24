# CLASSIFICATION: UNCLASSIFIED
.PHONY: build test lint vet docker-build clean

SERVICE_NAME := svc-radar-ingestion
IMAGE_TAG ?= latest

build:
go build ./...

test:
go test -race -count=1 ./...

test-integration:
go test -tags integration -race -count=1 ./...

lint:
go vet ./...

vet:
go vet ./...

docker-build:
docker build -t $(SERVICE_NAME):$(IMAGE_TAG) \
-f Dockerfile \
..

clean:
go clean ./...
