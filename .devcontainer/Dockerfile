ARG REFRESH_CACHE=2025-02-02
FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

# install dependecies for development by wails
RUN --mount=type=cache,target=/var/cache/apt \
    apt-get update && \
    apt-get install -y \
    libgtk-3-dev \
    libwebkit2gtk-4.0-dev \
    zenity \
    && rm -rf /var/lib/apt/lists/*

# install wails v2
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest