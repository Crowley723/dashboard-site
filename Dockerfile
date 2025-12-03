FROM node:24-alpine@sha256:2867d550cf9d8bb50059a0fff528741f11a84d985c732e60e19e8e75c7239c43 AS frontend-build
RUN npm install -g pnpm
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./

RUN pnpm install
COPY web/ ./
RUN pnpm run build

FROM golang:1.25-alpine@sha256:3587db7cc96576822c606d119729370dbf581931c5f43ac6d3fa03ab4ed85a10 AS backend-build
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
COPY internal/ ./internal/
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o dashboard-site ./main.go

FROM alpine:latest@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375 AS runtime
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1001 -S app && \
    adduser -u 1001 -S app -G app

WORKDIR /app
COPY --from=backend-build /app/dashboard-site .
COPY --from=frontend-build /app/web/dist ./web/dist

RUN chown -R app:app /app

USER app

EXPOSE 8080
CMD ["./dashboard-site",  "-c", "config.yaml"]