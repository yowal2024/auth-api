FROM node:24-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build -- --outDir /src/webdist --emptyOutDir

FROM golang:1.27.3-alpine AS go-base
WORKDIR /src/backend
ENV CGO_ENABLED=0
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./

FROM go-base AS test
CMD ["go", "test", "./...", "-count=1", "-v", "-timeout=180s"]

FROM go-base AS build
COPY --from=frontend /src/webdist ./internal/webassets/dist
RUN go build -trimpath -tags=embeddedweb -o /out/auth-api ./cmd/auth-api

FROM alpine:3.23 AS runtime
RUN apk add --no-cache ca-certificates tzdata && addgroup -S authapi && adduser -S -G authapi authapi && mkdir -p /app/runtime && chown authapi:authapi /app/runtime
COPY --from=build /out/auth-api /usr/local/bin/auth-api
COPY docker/config.yaml /app/config.yaml
WORKDIR /app/runtime
USER authapi
EXPOSE 8080
CMD ["sh", "-c", "test -f config.yaml || cp /app/config.yaml config.yaml; exec auth-api -config config.yaml -addr :8080"]
