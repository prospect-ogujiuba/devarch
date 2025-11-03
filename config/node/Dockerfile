# Node.js 20 LTS with comprehensive development tools
FROM node:20-bookworm-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    # Build tools
    build-essential \
    python3 \
    make \
    g++ \
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
    # Process manager dependencies
    dumb-init \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install global npm packages
RUN npm install -g \
    # Package managers
    pnpm \
    yarn \
    # Development servers
    nodemon \
    ts-node \
    ts-node-dev \
    # Build tools
    typescript \
    @types/node \
    webpack \
    webpack-cli \
    vite \
    esbuild \
    # Frameworks CLIs
    @nestjs/cli \
    @angular/cli \
    @vue/cli \
    create-react-app \
    create-next-app \
    express-generator \
    # Database tools
    typeorm \
    prisma \
    sequelize-cli \
    knex \
    # API tools
    @apollo/server \
    graphql \
    # Testing
    jest \
    mocha \
    @types/jest \
    # Linting & Formatting
    eslint \
    prettier \
    @typescript-eslint/eslint-plugin \
    @typescript-eslint/parser \
    # Process management
    pm2 \
    concurrently \
    # Utilities
    dotenv-cli \
    cross-env \
    rimraf \
    npm-check-updates

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
    chsh -s $(which zsh)

# Setup pnpm
RUN pnpm setup && \
    pnpm config set store-dir /root/.pnpm-store

# Create app directory
WORKDIR /app

# Create a non-root user
RUN useradd -m -s /bin/zsh node-user && \
    echo "node-user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R node-user:node-user /app

# Copy package files if they exist (for caching)
COPY package*.json pnpm-lock.yaml yarn.lock ./ 2>/dev/null || :

# Switch to non-root user
USER node-user

# Expose common Node.js application ports
EXPOSE 3000 3001 4000 5173 9229

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]

# Default command
CMD ["node", "--version"]
