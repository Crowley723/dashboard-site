export PATH := /home/brynn/.nvm/versions/node/v23.1.0/bin:$(PATH)
.PHONY: dev-frontend dev-backend dev install

install:
	go mod download
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm install"

dev-frontend:
	cd web && bash -c "source ~/.nvm/nvm.sh && pnpm run dev"

dev-backend:
	GO_ENV=development reflex -r '\.go$$' -s -- go run ./main.go -c config.yaml

dev:
	make -j2 dev-backend dev-frontend