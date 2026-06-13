FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    nodejs \
    npm \
    python3 \
    python3-pip \
    pipx \
    && rm -rf /var/lib/apt/lists/*
RUN userdel -r ubuntu 2>/dev/null; useradd -m -u 1000 -s /bin/bash agent
USER agent
WORKDIR /workspace
ENV PATH="/home/agent/.local/bin:${PATH}"
USER root
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*
USER agent
RUN mkdir -p /home/agent/.cache /home/agent/.local/state
