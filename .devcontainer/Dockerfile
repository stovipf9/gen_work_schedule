ARG REFRESH_CACHE=2025-02-02
FROM mcr.microsoft.com/devcontainers/go:2-1.25-trixie

# Yarn APT リポジトリを無効化（存在すれば削除）
RUN rm -f /etc/apt/sources.list.d/yarn.list

# install dependecies for development by wails
RUN --mount=type=cache,target=/var/cache/apt \
    apt-get update && \
    apt-get install -y \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    zenity \
    upx-ucl \
    && rm -rf /var/lib/apt/lists/*

# install wails v2
RUN go install github.com/wailsapp/wails/v2/cmd/wails@latest

# install go dev tools
# RUN go install golang.org/x/tools/cmd/goimports@0.38.0 \
#     && go install golang.org/x/tools/gopls@0.20.0

RUN go install github.com/vektra/mockery/v3@v3.6.1
RUN go install mvdan.cc/garble@latest