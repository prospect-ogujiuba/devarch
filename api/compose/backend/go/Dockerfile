# Go 1.22 dev image (super slim, non-brittle)
FROM golang:1.22-bookworm

# Install system deps & DB clients (like your PHP container, but minimal)
RUN apt-get update && apt-get install -y \
    sudo \
    gcc \
    g++ \
    make \
    cmake \
    default-mysql-client \
    postgresql-client \
    redis-tools \
    vim \
    nano \
    git \
    curl \
    wget \
    zsh \
    htop \
    ca-certificates \
    netcat-openbsd \
    libssl-dev \
    pkg-config \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Go env defaults
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Install Oh My Zsh non-interactively
RUN RUNZSH=no CHSH=no KEEP_ZSHRC=yes \
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
    chsh -s "$(which zsh)"

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

# App directory
WORKDIR /app

# Non-root user (same pattern as other stacks)
RUN useradd -m -s /bin/zsh go-user && \
    echo "go-user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R go-user:go-user /app /go

USER go-user

# Common Go service ports (app, admin, delve, pprof)
EXPOSE 8080 8081 2345 6060

# Default command (override in compose if needed)
CMD ["go", "version"]
