FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /devarch-import ./cmd/import

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /devarch-api /usr/local/bin/
COPY --from=builder /devarch-migrate /usr/local/bin/
COPY --from=builder /devarch-import /usr/local/bin/
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["devarch-api"]
