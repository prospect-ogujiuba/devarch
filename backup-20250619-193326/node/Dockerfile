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
    ts-node \
    serve

# Create app directory
WORKDIR /var/www/html

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# Set up directory permissions
RUN chown -R nextjs:nodejs /var/www/html

# Expose common Node.js ports
EXPOSE 3000 3001 8080 8000

# Start a simple HTTP server that can serve any directory
# This keeps the container running and ready to serve files
CMD ["serve", "-s", ".", "-p", "3000", "--cors"]