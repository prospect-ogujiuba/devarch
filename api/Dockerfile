FROM golang:1.23-alpine AS base

RUN apk add --no-cache git ca-certificates tzdata podman podman-compose && \
    mkdir -p /etc/containers && \
    printf '[engine]\nremote = true\n' > /etc/containers/containers.conf

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# --- Production build ---
FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-import ./cmd/import
RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-export ./cmd/export

FROM alpine:3.19 AS production

RUN apk add --no-cache ca-certificates tzdata podman podman-compose && \
    mkdir -p /etc/containers && \
    printf '[engine]\nremote = true\n' > /etc/containers/containers.conf

WORKDIR /app

COPY --from=builder /devarch-api /usr/local/bin/
COPY --from=builder /devarch-migrate /usr/local/bin/
COPY --from=builder /devarch-import /usr/local/bin/
COPY --from=builder /devarch-export /usr/local/bin/
COPY --from=builder /app/migrations ./migrations

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8080

CMD ["entrypoint.sh"]

# --- Dev with air hot reload ---
FROM base AS dev

RUN go install github.com/air-verse/air@v1.61.7
RUN CGO_ENABLED=0 go build -o /usr/local/bin/devarch-migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -o /usr/local/bin/devarch-import ./cmd/import
RUN CGO_ENABLED=0 go build -o /usr/local/bin/devarch-export ./cmd/export

COPY entrypoint-dev.sh /usr/local/bin/entrypoint-dev.sh
RUN chmod +x /usr/local/bin/entrypoint-dev.sh

EXPOSE 8080

CMD ["entrypoint-dev.sh"]
