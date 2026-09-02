FROM node:25-alpine@sha256:bdf2cca6fe3dabd014ea60163eca3f0f7015fbd5c7ee1b0e9ccb4ced6eb02ef4 AS frontend-build
RUN npm install -g pnpm
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./

RUN pnpm install
COPY web/ ./
RUN pnpm run build

FROM golang:1.27-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS backend-build
RUN apk add --no-cache git
WORKDIR /app

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

COPY go.mod go.sum ./
COPY internal/ ./internal/
COPY *.go ./
COPY VERSION ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X 'homelab-dashboard/internal/version.Version=${VERSION}' \
              -X 'homelab-dashboard/internal/version.GitCommit=${GIT_COMMIT}' \
              -X 'homelab-dashboard/internal/version.BuildTime=${BUILD_TIME}'" \
    -o dashboard-site ./main.go

FROM alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS runtime
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