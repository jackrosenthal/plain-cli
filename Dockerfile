FROM archlinux:latest

# Disable pacman's download sandbox (uses alpm UID not mapped in rootless containers)
RUN echo "DisableSandbox" >> /etc/pacman.conf

# Update and install system packages
RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
        git \
        go \
        golangci-lint \
        gofumpt \
        nodejs \
        npm \
        ca-certificates \
        base-devel && \
    pacman -Scc --noconfirm

# Go environment
ENV GOPATH=/root/go
ENV PATH=$PATH:/root/go/bin:/root/.local/bin

# Install Claude Code (native installer) and OpenAI Codex CLI
RUN curl -fsSL https://claude.ai/install.sh | bash
RUN npm install -g @openai/codex

# Configure git
RUN git config --global user.name "Jack Rosenthal" && \
    git config --global user.email "jack@rosenth.al" && \
    git config --global init.defaultBranch main && \
    git config --global safe.directory '*'

WORKDIR /workspace
