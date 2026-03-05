FROM golang:1.24-alpine AS base

RUN apk add --no-cache git ca-certificates podman curl py3-pip && \
    pip install --break-system-packages ramalama && \
    mkdir -p /etc/containers && \
    printf '[engine]\nremote = true\n' > /etc/containers/containers.conf

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# --- Production build ---
FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-ai ./cmd/server

FROM alpine:3.19 AS production

RUN apk add --no-cache ca-certificates podman curl py3-pip && \
    pip install --break-system-packages ramalama && \
    mkdir -p /etc/containers && \
    printf '[engine]\nremote = true\n' > /etc/containers/containers.conf

COPY --from=builder /devarch-ai /usr/local/bin/

EXPOSE 8080

CMD ["devarch-ai"]

# --- Dev with air hot reload ---
FROM base AS dev

RUN go install github.com/air-verse/air@v1.61.7

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
