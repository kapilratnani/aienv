FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    golang-go \
    nodejs \
    npm \
    python3 \
    python3-pip \
    pipx \
    && rm -rf /var/lib/apt/lists/*
RUN userdel -r ubuntu 2>/dev/null; useradd -m -u 1000 -s /bin/bash user
USER user
WORKDIR /workspace
ENV PATH="/home/user/.local/bin:${PATH}"
USER root
RUN npm i -g opencode-ai
USER user
