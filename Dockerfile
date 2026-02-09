FROM golang:1.22-bullseye AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o forum ./cmd/forum

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/forum /app/forum
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/schema.sql /app/schema.sql

ENV PORT=8080
EXPOSE 8080

CMD ["/app/forum"]
