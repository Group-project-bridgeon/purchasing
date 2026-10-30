# Go patch toolchain verified against the official release and image manifest.
FROM golang:1.27.3-alpine AS builder
WORKDIR /src
ENV GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /out/purchasing ./cmd/server

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -g 10001 -S appgroup \
 && adduser -u 10001 -S -G appgroup appuser
WORKDIR /app
COPY --from=builder /out/purchasing /app/purchasing
USER 10001:10001
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 CMD wget -qO- "http://127.0.0.1:${SERVER_PORT:-8080}/health" >/dev/null || exit 1
ENTRYPOINT ["/app/purchasing"]
