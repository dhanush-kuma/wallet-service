# ---------- BUILDER ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# install dependencies needed for go modules
RUN apk add --no-cache git curl

# install migrate binary
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
  | tar xvz && mv migrate /usr/local/bin/migrate

# cache go modules
COPY go.mod go.sum ./
RUN go mod download

# copy source
COPY . .

# build binary
RUN go build -o wallet-service ./cmd/server


# ---------- RUNTIME ----------
FROM alpine:latest

WORKDIR /app

# needed for HTTPS + seed execution
RUN apk add --no-cache ca-certificates postgresql-client

# copy binary and tools
COPY --from=builder /app/wallet-service .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# copy migrations and startup script
COPY migrations ./migrations
COPY scripts/start.sh ./start.sh

RUN chmod +x start.sh

EXPOSE 8080

CMD ["./start.sh"]