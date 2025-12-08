# Vite development server for Django multi-app asset compilation
FROM node:22-bookworm

# Install supervisor for multi-app orchestration
RUN apt-get update && apt-get install -y \
    supervisor \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /apps

# Install global packages
RUN npm install -g vite

# Create non-root user
RUN groupadd -g 1000 node || true && \
    usermod -u 1000 -g 1000 node || useradd -m -u 1000 -g 1000 node

# Set ownership
RUN chown -R node:node /apps

# Switch to non-root user
USER node

# Expose Vite dev server ports (5173-5183 for up to 10 apps)
EXPOSE 5173 5174 5175 5176 5177 5178 5179 5180 5181 5182 5183

# Default command (overridden by auto-discovery script)
CMD ["vite", "--version"]
