# --- Build stage ---
FROM golang:1.22-alpine AS builder

# Tillåt Go att auto-ladda rätt toolchain enligt go.mod
ENV GOTOOLCHAIN=auto

WORKDIR /app

# Kopiera mod-filer först (bättre cache)
COPY go.mod go.sum ./
RUN go mod download

# Kopiera resten av koden
COPY . .

# Bygg binären för api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api-gateway ./cmd/api-gateway

# --- Runtime stage ---
FROM gcr.io/distroless/base-debian12

WORKDIR /
COPY --from=builder /bin/api-gateway /api-gateway

EXPOSE 8088
USER nonroot:nonroot

ENTRYPOINT ["/api-gateway"]
