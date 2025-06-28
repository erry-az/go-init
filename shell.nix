{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Go and tools
    go
    gopls
    golangci-lint
    gotools
    go-tools # Instead of go-mockgen
    delve
    
    # Database tools
    postgresql
    goose
    sqlc
    
    # Protobuf tools
    protobuf
    buf
    
    # Docker and container tools
    docker
    docker-compose
    
    # Terminal tools
    starship
    fish # Using fish shell
    fzf # For enhanced history search
  ];
  
  shellHook = ''
    export GOPATH="$HOME/go"
    export PATH="$GOPATH/bin:$PATH"
    
    # Create fish config directory if it doesn't exist
    mkdir -p ~/.config/fish
    
    # Configure fish to use starship
    echo 'starship init fish | source' > ~/.config/fish/config.fish
    
    # Start fish shell with starship
    exec fish -l
  '';
}
