export PATH := /home/brynn/.local/share/pnpm:$(PATH)
.PHONY: dev-frontend dev-backend dev install

install:
	go mod download
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm install"

dev-frontend:
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm run dev"

dev-backend:
	GO_ENV=development reflex -r '\.go$$' -s -- go run ./main.go -c config.yaml


dev-backend-debug:
	GO_ENV=development reflex -r '\.go$$' -s -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient ./main.go -- -c config.yaml

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