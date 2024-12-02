FROM mcr.microsoft.com/devcontainers/universal:2
COPY .devcontainer/.bash_aliases /home/codespace/.bash_aliases
RUN set -eax; \
    sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
