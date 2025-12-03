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
    # Mail relay
    msmtp \
    msmtp-mta \
    # Process manager dependencies
    dumb-init \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install global npm packages (no yarn here – Node 20 has Corepack/Yarn already)
RUN npm install -g \
    # Package managers
    pnpm \
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

# Configure msmtp for mail relay
RUN echo "account default" > /etc/msmtprc \
    && echo "host mailpit" >> /etc/msmtprc \
    && echo "port 1025" >> /etc/msmtprc \
    && echo "from noreply@devarch.test" >> /etc/msmtprc \
    && chmod 644 /etc/msmtprc

# Install Oh My Zsh non-interactively for root
RUN RUNZSH=no CHSH=no KEEP_ZSHRC=yes \
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
    chsh -s "$(which zsh)"

# Create app directory
WORKDIR /app

# Create a non-root user
RUN useradd -m -s /bin/zsh node-user && \
    echo "node-user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R node-user:node-user /app

# Switch to non-root user
USER node-user

# Configure pnpm store for node-user (no pnpm setup needed)
RUN pnpm config set store-dir /home/node-user/.pnpm-store

# (Optional) if you want a separate Oh My Zsh for node-user, you could do:
# RUN RUNZSH=no CHSH=no KEEP_ZSHRC=yes \
#     sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

# Expose common Node.js application ports
EXPOSE 3000 3001 4000 5173 9229

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]

# Default command
CMD ["node", "--version"]
