# Python 3.12 with comprehensive development tools
FROM python:3.12-slim-bookworm

# Base environment
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    POETRY_VERSION=1.7.1 \
    POETRY_HOME="/opt/poetry" \
    VIRTUAL_ENV="/opt/venv"

# Put Poetry and the venv first on PATH
ENV PATH="$POETRY_HOME/bin:$VIRTUAL_ENV/bin:$PATH"

# Install system dependencies and development tools
RUN apt-get update && apt-get install -y \
    # Build essentials
    build-essential \
    gcc \
    g++ \
    make \
    cmake \
    pkg-config \
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
    libmariadb-dev-compat \
    libmariadb-dev \
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

# 🔹 Create a virtual environment (this is what makes pip happy)
RUN python -m venv "$VIRTUAL_ENV"

# Install Poetry into the (system) Python – fine, it drops binaries into /opt/poetry/bin
RUN curl -sSL https://install.python-poetry.org | python3 -

# 🔹 All pip installs from here are inside the venv
RUN pip install --upgrade pip setuptools wheel && \
    pip install \
        django \
        flask \
        fastapi \
        uvicorn[standard] \
        gunicorn \
        celery[redis] \
        flower \
        sqlalchemy \
        alembic \
        django-extensions \
        djangorestframework \
        flask-restful \
        pydantic \
        pytest \
        pytest-django \
        pytest-asyncio \
        pytest-cov \
        ipython \
        jupyter \
        jupyterlab \
        black \
        flake8 \
        mypy \
        pylint \
        bandit \
        debugpy \
        django-debug-toolbar \
        psycopg2-binary \
        pymongo \
        redis \
        mysqlclient \
        httpx \
        requests \
        aiohttp \
        pandas \
        numpy \
        rq \
        dramatiq[redis] \
        python-dotenv \
        pyyaml \
        click

# Install Oh My Zsh and set Zsh as default shell
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" \
    && chsh -s $(which zsh)

# Create app directory
WORKDIR /app

# Create a non-root user
RUN useradd -m -s /bin/zsh python && \
    echo "python ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    chown -R python:python /app

# Switch to non-root user
USER python

# Expose common Python application ports
EXPOSE 8000 5000 5555 8888

# Default command (can be overridden)
CMD ["python", "--version"]
