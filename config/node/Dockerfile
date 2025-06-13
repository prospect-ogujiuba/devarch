FROM node:lts-alpine

# Install system dependencies
RUN apk add --no-cache \
    git \
    curl \
    bash \
    zsh \
    vim \
    nano

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Install global development tools
RUN npm install -g \
    nodemon \
    pm2 \
    @nestjs/cli \
    create-react-app \
    create-next-app \
    express-generator \
    typescript \
    ts-node

# Create app directory
WORKDIR /app

# Copy smart entrypoint script
COPY smart-entrypoint.sh /usr/local/bin/smart-entrypoint.sh
RUN chmod +x /usr/local/bin/smart-entrypoint.sh

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# Expose common Node.js ports
EXPOSE 3000 3001 8080 8000

# Use smart entrypoint
ENTRYPOINT ["/usr/local/bin/smart-entrypoint.sh"]

# Default command
CMD ["npm", "start"]