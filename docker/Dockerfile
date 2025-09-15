FROM node:22-alpine AS frontend-build
RUN npm install -g pnpm
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./

RUN pnpm install
COPY web/ ./
RUN pnpm run build

FROM golang:1.25-alpine AS backend-build
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
COPY internal/ ./internal/
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o dashboard-site ./main.go

FROM alpine:latest AS runtime
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1001 -S app && \
    adduser -u 1001 -S app -G app

WORKDIR /app
COPY --from=backend-build /app/dashboard-site .
COPY --from=frontend-build /app/web/dist ./web/dist

COPY config.yaml ./
RUN chown -R app:app /app

USER app

EXPOSE 8080
CMD ["./dashboard-site",  "-c", "config.yaml"]