export PATH := /home/brynn/.local/share/pnpm:$(PATH)
.PHONY: dev-frontend dev-backend dev install dev-backend-debug dev-debug

install:
	go mod download
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm install"

dev-frontend:
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm run dev"

dev-backend:
	GO_ENV=development reflex -r '\.go$$' -s -- go run ./main.go -c config.docker.yaml

dev-backend-debug:
	GO_ENV=development reflex -r '\.go$$' -s -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient ./main.go -- -c config.docker.yaml

dev:
	@echo "Starting development servers..."
	@trap 'kill 0' INT TERM EXIT; \
	($(MAKE) dev-backend) & \
	($(MAKE) dev-frontend) & \
	wait

dev-debug:
	@echo "Starting development servers with debug..."
	@trap 'kill 0' INT TERM EXIT; \
	($(MAKE) dev-backend-debug) & \
	($(MAKE) dev-frontend) & \
	wait

build:
	docker build -t dashboard-site:latest -f docker/Dockerfile .

# Docker development commands
dev-docker:
	@echo "Starting Docker development environment..."
	cd docker && docker compose up --build

dev-docker-debug:
	@echo "Starting Docker development environment with debugger..."
	cd docker && docker compose run --rm --service-ports dashboard-app /usr/local/bin/start-debug.sh

dev-docker-down:
	@echo "Stopping Docker development environment..."
	cd docker && docker compose down

dev-docker-logs:
	@echo "Following Docker development logs..."
	cd docker && docker compose logs -f

dev-docker-rebuild:
	@echo "Rebuilding Docker development environment..."
	cd docker && docker compose down && docker compose up --build

TEST_FLAGS ?=

test:
	go test $(TEST_FLAGS) ./...

coverage:
	go test $(TEST_FLAGS) -cover ./...

coverage-html:
	go test -coverprofile=coverage.out $(TEST_FLAGS) ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

mocks:
	go generate ./...