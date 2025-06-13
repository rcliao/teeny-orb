# Execution container for teeny-orb sessions
FROM alpine:latest

# Install common development tools
RUN apk add --no-cache \
    bash \
    git \
    curl \
    wget \
    vim \
    nano \
    build-base \
    python3 \
    py3-pip \
    nodejs \
    npm

# Create workspace directory
RUN mkdir -p /workspace
WORKDIR /workspace

# Add a non-root user for security
RUN adduser -D -s /bin/bash workspace
USER workspace

# Set environment variables
ENV TERM=xterm-256color
ENV EDITOR=vim

# Keep container running
CMD ["tail", "-f", "/dev/null"]