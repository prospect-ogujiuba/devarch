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

# Install global Python tools and a simple HTTP server
RUN pip install --no-cache-dir \
    pipenv \
    poetry \
    virtualenv \
    uvicorn \
    gunicorn \
    pytest \
    black \
    flake8 \
    mypy \
    fastapi \
    flask \
    django \
    streamlit

# Create app directory
WORKDIR /var/www/html

# Create non-root user
RUN useradd --create-home --shell /bin/bash app
RUN chown -R app:app /var/www/html

# Expose common Python ports
EXPOSE 8000 8080 5000 3000

# Run a simple HTTP server that serves the current directory
# This keeps the container running and ready to serve files
CMD ["python", "-m", "http.server", "8000"]