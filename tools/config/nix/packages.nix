{ pkgs }:
{
  # Development packages organized by category
  packages = with pkgs; [
    # Go development tools - use specific version to avoid conflicts
    go_1_24
    gopls
    golangci-lint
    gotools
    go-tools
    delve

    # Database tools
    postgresql
    atlas
    sqlc

    # Protobuf tools
    protobuf
    buf

    # Container tools
    docker
    docker-compose

    # Shell and CLI tools
    zsh
    starship
    git
    fzf
    eza
    zoxide

    # ZSH plugins
    zsh-fzf-tab
    zsh-fast-syntax-highlighting
    zsh-completions
    zsh-autosuggestions
  ];
}