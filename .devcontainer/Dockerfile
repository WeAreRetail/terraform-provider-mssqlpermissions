# Inspiration: https://github.com/devcontainers/templates/blob/main/src/go-postgres/.devcontainer/Dockerfile
ARG GO_TAG=2.0-bookworm
FROM mcr.microsoft.com/devcontainers/go:${GO_TAG}

# Tool versions
ARG REMOTE_USER=vscode
ARG DELVE_VERSION=1.25.2
ARG GOLANGCI_LINT_VERSION=v2.5.0
ARG GORELEASER_VERSION=v2.11.2
ARG GOPLS_VERSION=latest
ARG STATICCHECK_VERSION=latest
ARG SHFMT_VERSION=latest
ARG GO_MOD_OUTDATED_VERSION=latest
ARG GOIMPORTS_VERSION=latest

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive \
    GOPATH=/go \
    PATH=/go/bin:$PATH

# Install additional packages
RUN apt-get update \
    && apt-get -y install --no-install-recommends procps lsb-release unzip gpg shellcheck curl jq \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/library-scripts

# Install Customizations
COPY customization ./customization
RUN bash ./customization/install_eza.sh \
    && bash ./customization/install_starship.sh \
    && bash ./customization/install_task.sh \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/library-scripts && rm -rf ./customization

# Install SQL Tools: SQLPackage and sqlcmd
COPY mssql/installSQLtools.sh installSQLtools.sh
RUN bash ./installSQLtools.sh \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/library-scripts && rm -rf ./installSQLtools.sh

# Configure terraformrc
COPY terraform/.terraformrc /root/.terraformrc

# Export bash history
RUN \
    SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
    && mkdir /commandhistory \
    && touch /commandhistory/.bash_history \
    && chown -R $REMOTE_USER /commandhistory \
    && echo "$SNIPPET" >> "/home/$REMOTE_USER/.bashrc"

# Install Go tools
RUN \
    # --> Delve for debugging
    go install github.com/go-delve/delve/cmd/dlv@v${DELVE_VERSION} \
    # --> Go language server
    && go install golang.org/x/tools/gopls@${GOPLS_VERSION} \
    # --> GolangCI-lint (updated installation method)
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin ${GOLANGCI_LINT_VERSION} \
    # --> Staticcheck
    && go install honnef.co/go/tools/cmd/staticcheck@${STATICCHECK_VERSION} \
    # --> GoReleaser
    && go install github.com/goreleaser/goreleaser/v2@${GORELEASER_VERSION} \
    # --> shfmt
    && go install mvdan.cc/sh/v3/cmd/shfmt@${SHFMT_VERSION} \
    # --> go-mod-outdated for dependency management
    && go install github.com/psampaz/go-mod-outdated@${GO_MOD_OUTDATED_VERSION} \
    # --> goimports for code formatting
    && go install golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION} \
    #
    # Clean up
    && rm -rf /go/src/ && rm -rf /go/pkg

# Switching to the user that will be used in the Dev Container
USER ${REMOTE_USER}

COPY terraform/.terraformrc /tmp/.terraformrc

# Customize bash
RUN \
    # Add alias and PATH
    echo "alias tf=terraform" >> "$HOME/.bashrc" \
    && echo "alias tfa='terraform apply'" >> "$HOME/.bashrc" \
    && echo "alias tfp='terraform plan'" >> "$HOME/.bashrc" \
    && echo "alias tfd='terraform destroy'" >> "$HOME/.bashrc" \
    && echo "alias ll=\"eza -la --icons\"" >> "$HOME/.bashrc" \
    && echo "alias la=\"eza -la --icons\"" >> "$HOME/.bashrc" \
    && echo "alias ls=\"eza --icons\"" >> "$HOME/.bashrc" \
    && echo 'eval "$(starship init bash)"' >> "$HOME/.bashrc" \
    && echo 'eval "$(task --completion bash)"' >> "$HOME/.bashrc" \
    && echo "export PATH=$PATH:/usr/bin/"  >> "$HOME/.bashrc" \
    && cp "/tmp/.terraformrc" "$HOME/.terraformrc"

USER root

# Final cleanup and set proper permissions
RUN chown -R ${REMOTE_USER}:${REMOTE_USER} /home/${REMOTE_USER} \
    && chmod -R 755 /home/${REMOTE_USER}

# Switch back to dialog for any ad-hoc use of apt-get
ENV DEBIAN_FRONTEND=dialog

# Set the default user
USER ${REMOTE_USER}

# Set the default working directory
WORKDIR /workspace

# Add metadata labels
LABEL maintainer="Terraform Provider MSSQL Permissions Dev Team" \
      version="1.0" \
      description="Development container for terraform-provider-mssqlpermissions"

# Add health check for the development environment
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD go version || exit 1
