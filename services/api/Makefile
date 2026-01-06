.PHONY: build run test clean importer import-photos import-posko import-all

# Build targets
build:
	go build -o bin/api ./cmd/api

importer:
	go build -o bin/importer ./cmd/importer

build-all: build importer

# Run targets
run:
	go run ./cmd/api

# Import commands
import-photos: importer
	./bin/importer -photos -verbose

import-photos-dry: importer
	./bin/importer -photos -dry-run -verbose

import-posko: importer
	./bin/importer -posko -verbose

import-posko-dry: importer
	./bin/importer -posko -dry-run -verbose

import-all: importer
	./bin/importer -all -verbose

import-all-dry: importer
	./bin/importer -all -dry-run -verbose

# Test targets
test:
	go test -v ./...

# Clean
clean:
	rm -rf bin/

# Docker
docker-build:
	docker build -t senyar-api .

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build API binary"
	@echo "  importer       - Build importer CLI tool"
	@echo "  build-all      - Build all binaries"
	@echo "  run            - Run API server"
	@echo "  import-photos  - Import uncached photos from ODK"
	@echo "  import-posko   - Sync posko data from ODK"
	@echo "  import-all     - Sync all data and photos"
	@echo "  *-dry          - Dry run versions (no changes)"
	@echo "  test           - Run tests"
	@echo "  clean          - Remove build artifacts"
