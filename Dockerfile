FROM node:24-alpine@sha256:6085692e50e15e1a4e83f0452e56016ea3852d27754a87365372d85f4012a898 AS frontend-build
RUN npm install -g pnpm
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./

RUN pnpm install
COPY web/ ./
RUN pnpm run build

FROM golang:1.25-alpine@sha256:26111811bc967321e7b6f852e914d14bede324cd1accb7f81811929a6a57fea9 AS backend-build
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