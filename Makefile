.PHONY: all build-backend build-frontend clean test

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

all: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	go build -ldflags "-X 'github.com/Progressio-dev/grafana-reporter/pkg/plugin.BuildVersion=$(VERSION)' -X 'github.com/Progressio-dev/grafana-reporter/pkg/plugin.BuildTime=$(BUILD_TIME)'" -o dist/gpx_grafana-reporter ./pkg

build-frontend:
	@echo "Building frontend..."
	npm run build

clean:
	@echo "Cleaning..."
	rm -rf dist/
	rm -rf node_modules/

test:
	@echo "Running tests..."
	go test ./...
	npm run test:ci

install-deps:
	@echo "Installing dependencies..."
	go mod download
	npm install

dev-backend:
	@echo "Running backend in dev mode..."
	go run ./pkg

dev-frontend:
	@echo "Running frontend in dev mode..."
	npm run dev

lint:
	@echo "Linting..."
	npm run lint
