FROM jc21/nginx-proxy-manager:latest

# Install required packages
RUN apt-get update && apt-get install -y \
    zsh \
    git \
    curl \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install Oh My Zsh and set Zsh as default shell
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" \
    && chsh -s $(which zsh)