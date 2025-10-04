export PATH := /home/brynn/.local/share/pnpm:$(PATH)
.PHONY: dev-frontend dev-backend dev install

install:
	go mod download
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm install"

dev-frontend-bare:
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm run dev"

dev-backend-bare:
	GO_ENV=development reflex -r '\.go$$' -s -- go run ./main.go -c config.yaml


dev-backend-debug-bare:
	GO_ENV=development reflex -r '\.go$$' -s -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient ./main.go -- -c config.yaml

dev-bare:
	@echo "Starting development servers..."
	@trap 'kill 0' INT TERM EXIT; \
	($(MAKE) dev-backend) & \
	($(MAKE) dev-frontend) & \
	wait

dev-debug-bare:
	@echo "Starting development servers with debug..."
	@trap 'kill 0' INT TERM EXIT; \
	($(MAKE) dev-backend-debug) & \
	($(MAKE) dev-frontend) & \
	wait

build:
	docker build -t dashboard-site:latest -f docker/Dockerfile .

# Docker development commands
dev:
	@echo "Starting Docker development environment..."
	cd docker && docker compose up --build

dev-debug:
	@echo "Starting Docker development environment with debugger..."
	cd docker && docker compose run --rm --service-ports dashboard-app /usr/local/bin/start-debug.sh

dev-stop:
	@echo "Stopping Docker development environment..."
	cd docker && docker compose down

dev-logs:
	@echo "Following Docker development logs..."
	cd docker && docker compose logs -f

dev-rebuild:
	@echo "Rebuilding Docker development environment..."
	cd docker && docker compose down && docker compose up --build

coverage:
	go test -cover -v ./...

coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

mocks:
	go generate ./...