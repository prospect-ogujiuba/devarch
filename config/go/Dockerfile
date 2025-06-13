FROM golang:1.21-alpine AS builder

# Install system dependencies
RUN apk add --no-cache \
    git \
    curl \
    bash \
    zsh \
    vim \
    nano \
    build-base

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Install Go development tools (compatible versions for Go 1.21)
RUN go install github.com/cosmtrek/air@v1.49.0 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.21.2 && \
    go install golang.org/x/tools/cmd/goimports@v0.13.0 && \
    go install honnef.co/go/tools/cmd/staticcheck@2023.1.6

# Create app directory
WORKDIR /app

# Copy smart entrypoint script
COPY smart-entrypoint.sh /usr/local/bin/smart-entrypoint.sh
RUN chmod +x /usr/local/bin/smart-entrypoint.sh

# Create non-root user
RUN adduser -D -s /bin/bash app
RUN chown -R app:app /app

# Expose common Go ports
EXPOSE 8080 8000 3000 9000

# Use smart entrypoint
ENTRYPOINT ["/usr/local/bin/smart-entrypoint.sh"]

# Default command
CMD ["go", "run", "main.go"]