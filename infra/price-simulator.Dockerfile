FROM golang:1.22-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/price-simulator ./cmd/price-simulator

FROM gcr.io/distroless/base-debian12

WORKDIR /
COPY --from=builder /bin/price-simulator /price-simulator

USER nonroot:nonroot
ENTRYPOINT ["/price-simulator"]
