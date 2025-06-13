FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    zsh \
    vim \
    nano \
    build-essential \
    libpq-dev \
    libmariadb-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Install global Python tools
RUN pip install --no-cache-dir \
    pipenv \
    poetry \
    virtualenv \
    uvicorn \
    gunicorn \
    pytest \
    black \
    flake8 \
    mypy

# Create app directory
WORKDIR /app

# Copy smart entrypoint script
COPY smart-entrypoint.sh /usr/local/bin/smart-entrypoint.sh
RUN chmod +x /usr/local/bin/smart-entrypoint.sh

# Create non-root user
RUN useradd --create-home --shell /bin/bash app
RUN chown -R app:app /app

# Expose common Python ports
EXPOSE 8000 8080 5000 3000

# Use smart entrypoint
ENTRYPOINT ["/usr/local/bin/smart-entrypoint.sh"]

# Default command
CMD ["python", "main.py"]