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

# Install Go development tools
RUN go install github.com/cosmtrek/air@v1.49.0 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.21.2 && \
    go install golang.org/x/tools/cmd/goimports@v0.13.0 && \
    go install honnef.co/go/tools/cmd/staticcheck@2023.1.6

# Create a simple Go file server
RUN mkdir -p /app
WORKDIR /app

# Create a simple file server that serves the /var/www/html directory
RUN cat > main.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Serve files from /var/www/html
    fs := http.FileServer(http.Dir("/var/www/html"))
    http.Handle("/", fs)

    fmt.Printf("Go file server running on port %s\n", port)
    fmt.Println("Serving files from /var/www/html")
    
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
EOF

# Build the file server
RUN go mod init fileserver && go build -o fileserver main.go

# Create app directory
WORKDIR /var/www/html

# Create non-root user
RUN adduser -D -s /bin/bash app
RUN chown -R app:app /var/www/html

# Expose common Go ports
EXPOSE 8080 8000 3000 9000

# Run the Go file server
CMD ["/app/fileserver"]