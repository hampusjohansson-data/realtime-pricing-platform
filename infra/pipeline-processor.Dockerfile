FROM golang:1.22-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/pipeline-processor ./cmd/pipeline-processor

FROM gcr.io/distroless/base-debian12

WORKDIR /
COPY --from=builder /bin/pipeline-processor /pipeline-processor

USER nonroot:nonroot
ENTRYPOINT ["/pipeline-processor"]
