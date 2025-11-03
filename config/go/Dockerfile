# Go 1.22 with comprehensive development tools
FROM golang:1.22-bookworm

# Install system dependencies and tools
RUN apt-get update && apt-get install -y \
    # Build essentials
    build-essential \
    gcc \
    g++ \
    make \
    cmake \
    # Database clients
    postgresql-client \
    default-mysql-client \
    redis-tools \
    # Network tools
    curl \
    wget \
    netcat-openbsd \
    # Development tools
    git \
    vim \
    nano \
    zsh \
    sudo \
    htop \
    # Additional libraries
    libssl-dev \
    pkg-config \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools and packages
RUN go install -v \
    # Development tools
    github.com/cosmtrek/air@latest \
    github.com/go-delve/delve/cmd/dlv@latest \
    github.com/swaggo/swag/cmd/swag@latest \
    github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
    # Code quality
    golang.org/x/tools/gopls@latest \
    golang.org/x/lint/golint@latest \
    github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    honnef.co/go/tools/cmd/staticcheck@latest \
    github.com/securego/gosec/v2/cmd/gosec@latest \
    # Testing
    github.com/rakyll/gotest@latest \
    github.com/kyoh86/richgo@latest \
    github.com/onsi/ginkgo/v2/ginkgo@latest \
    # Database tools
    github.com/pressly/goose/v3/cmd/goose@latest \
    entgo.io/ent/cmd/ent@latest \
    # API tools
    github.com/99designs/gqlgen@latest \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest \
    google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    # Utilities
    github.com/spf13/cobra-cli@latest \
    github.com/google/wire/cmd/wire@latest

# Install Protocol Buffers compiler
RUN PROTOC_VERSION=25.1 && \
    curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
    rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
    chsh -s $(which zsh)

# Setup Go workspace
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPRIVATE="" \
    GOPROXY=https://proxy.golang.org,direct

# Create app directory
WORKDIR /app

# Create a non-root user
RUN useradd -m -s /bin/zsh go-user && \
    echo "go-user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R go-user:go-user /app && \
    chown -R go-user:go-user /go

# Pre-download common Go dependencies for caching
RUN go mod download \
    github.com/gin-gonic/gin \
    github.com/labstack/echo/v4 \
    github.com/gofiber/fiber/v2 \
    github.com/gorilla/mux \
    gorm.io/gorm \
    gorm.io/driver/postgres \
    gorm.io/driver/mysql \
    github.com/go-redis/redis/v8 \
    go.mongodb.org/mongo-driver \
    || true

# Switch to non-root user
USER go-user

# Expose common Go application ports
EXPOSE 8080 8081 2345 6060

# Default command
CMD ["go", "version"]
