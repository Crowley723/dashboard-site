FROM node:25-alpine@sha256:ad82ecad30371c43f4057aaa4800a8ed88f9446553a2d21323710c7b937177fc AS frontend-build
RUN npm install -g pnpm
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./

RUN pnpm install
COPY web/ ./
RUN pnpm run build

FROM golang:1.26-alpine@sha256:27f829349da645e287cb195a9921c106fc224eeebbdc33aeb0f4fca2382befa6 AS backend-build
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

FROM alpine:latest@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659 AS runtime
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