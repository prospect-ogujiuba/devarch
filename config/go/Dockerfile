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

# Create the Go file server using echo instead of heredoc
RUN echo 'package main' > main.go && \
    echo '' >> main.go && \
    echo 'import (' >> main.go && \
    echo '    "fmt"' >> main.go && \
    echo '    "log"' >> main.go && \
    echo '    "net/http"' >> main.go && \
    echo '    "os"' >> main.go && \
    echo ')' >> main.go && \
    echo '' >> main.go && \
    echo 'func main() {' >> main.go && \
    echo '    port := os.Getenv("PORT")' >> main.go && \
    echo '    if port == "" {' >> main.go && \
    echo '        port = "8080"' >> main.go && \
    echo '    }' >> main.go && \
    echo '' >> main.go && \
    echo '    // Serve files from /var/www/html' >> main.go && \
    echo '    fs := http.FileServer(http.Dir("/var/www/html"))' >> main.go && \
    echo '    http.Handle("/", fs)' >> main.go && \
    echo '' >> main.go && \
    echo '    fmt.Printf("Go file server running on port %s\\n", port)' >> main.go && \
    echo '    fmt.Println("Serving files from /var/www/html")' >> main.go && \
    echo '    ' >> main.go && \
    echo '    log.Fatal(http.ListenAndServe(":"+port, nil))' >> main.go && \
    echo '}' >> main.go

# Build the file server
RUN go mod init fileserver && go build -o /usr/local/bin/fileserver main.go

# Verify the binary exists and is executable
RUN ls -la /usr/local/bin/fileserver && chmod +x /usr/local/bin/fileserver

# Create app directory
WORKDIR /var/www/html

# Create non-root user
RUN adduser -D -s /bin/bash app
RUN chown -R app:app /var/www/html

# Expose common Go ports
EXPOSE 8080 8000 3000 9000

# Run the Go file server
CMD ["fileserver"]