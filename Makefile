export PATH := /home/brynn/.local/share/pnpm:$(PATH)
.PHONY: dev-frontend dev-backend dev install dev-backend-debug dev-debug dev-local dev-local-debug dev-local-stop dev-cluster dev-cluster-delete dev-cluster-reset dev-cluster-status

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

dev-local:
	@echo "Starting local development against k3d cluster..."
	@if ! k3d cluster list | grep -q "^conduit-dev$$"; then \
	   echo "Cluster not found. Creating it..."; \
	   $(MAKE) dev-cluster; \
	fi
	@trap 'kill 0' INT TERM EXIT; \
	($(MAKE) dev-backend) & \
	($(MAKE) dev-frontend) & \
	wait

dev-local-debug:
	@echo "Starting local development with debugger against k3d cluster..."
	@if ! k3d cluster list | grep -q "^conduit-dev$$"; then \
	   echo "Cluster not found. Creating it..."; \
	   $(MAKE) dev-cluster; \
	fi
	@trap 'kill 0' INT TERM EXIT; \
	($(MAKE) dev-backend-debug) & \
	($(MAKE) dev-frontend) & \
	wait

dev-local-stop:
	@echo "Stopping local development servers..."
	@pkill -f "go run ./main.go" || true
	@pkill -f "pnpm run dev" || true
	@pkill -f "reflex" || true
	@echo "Stopped"

dev-cluster:
	@echo "Setting up local development cluster..."
	@./scripts/setup-dev-cluster.sh

dev-cluster-delete:
	@echo "Deleting local development cluster..."
	@k3d cluster delete conduit-dev

dev-cluster-reset: dev-cluster-delete dev-cluster
	@echo "Development cluster reset complete"

dev-cluster-status:
	@echo "Checking cluster status..."
	@k3d cluster list | grep conduit-dev && \
	   kubectl cluster-info --context k3d-conduit-dev || \
	   echo "Cluster not running. Run 'make dev-cluster' to create it."

build:
	docker build -t dashboard-site:latest -f docker/Dockerfile .

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