# Python 3.12 with comprehensive development tools
FROM python:3.12-slim-bookworm

# Set Python environment variables
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    POETRY_VERSION=1.7.1 \
    POETRY_HOME="/opt/poetry" \
    POETRY_VIRTUALENVS_CREATE=false \
    PATH="$POETRY_HOME/bin:$PATH"

# Install system dependencies and development tools
RUN apt-get update && apt-get install -y \
    # Build essentials
    build-essential \
    gcc \
    g++ \
    make \
    cmake \
    # Python development
    python3-dev \
    python3-pip \
    python3-venv \
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
    # Libraries for Python packages
    libpq-dev \
    libmysqlclient-dev \
    libssl-dev \
    libffi-dev \
    libxml2-dev \
    libxslt1-dev \
    libjpeg-dev \
    libpng-dev \
    libwebp-dev \
    zlib1g-dev \
    # Additional tools
    supervisor \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Poetry
RUN curl -sSL https://install.python-poetry.org | python3 -

# Install global Python packages
RUN pip install --upgrade pip setuptools wheel && \
    pip install \
    # Web frameworks
    django \
    flask \
    fastapi \
    uvicorn[standard] \
    gunicorn \
    # Async workers
    celery[redis] \
    flower \
    # Database ORMs
    sqlalchemy \
    alembic \
    django-extensions \
    # API tools
    djangorestframework \
    flask-restful \
    pydantic \
    # Testing
    pytest \
    pytest-django \
    pytest-asyncio \
    pytest-cov \
    # Development tools
    ipython \
    jupyter \
    jupyterlab \
    black \
    flake8 \
    mypy \
    pylint \
    bandit \
    # Debugging
    debugpy \
    django-debug-toolbar \
    # Database drivers
    psycopg2-binary \
    pymongo \
    redis \
    mysqlclient \
    # HTTP clients
    httpx \
    requests \
    aiohttp \
    # Data processing
    pandas \
    numpy \
    # Task queue
    rq \
    dramatiq[redis] \
    # Other utilities
    python-dotenv \
    pyyaml \
    click

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" && \
    chsh -s $(which zsh)

# Create app directory
WORKDIR /app

# Create a non-root user
RUN useradd -m -s /bin/zsh python && \
    echo "python ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R python:python /app

# Copy supervisor config if needed
COPY ../../config/supervisord/supervisord.conf /etc/supervisor/conf.d/supervisord.conf 2>/dev/null || :

# Switch to non-root user
USER python

# Expose common Python application ports
EXPOSE 8000 5000 5555 8888

# Default command (can be overridden)
CMD ["python", "--version"]
