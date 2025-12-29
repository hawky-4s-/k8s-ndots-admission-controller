.PHONY: build test lint docker-build

IMG ?= ndots-webhook:latest

# Build binary
build:
	go build -o bin/webhook ./cmd/webhook

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run linter
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t $(IMG) .

# Generate TLS certs for local dev
certs:
	./scripts/generate-certs.sh

# Create kind cluster
kind-create:
	kind create cluster --name ndots-dev

# Load image to kind
kind-load:
	kind load docker-image $(IMG) --name ndots-dev

# Deploy to kind
deploy:
	kubectl apply -f deploy/kubernetes/
