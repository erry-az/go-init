{ pkgs ? import <nixpkgs> {} }:

let
  # Import modular configuration
  environment = import ./tools/config/nix/environment.nix { inherit pkgs; };
  zshConfig = import ./tools/config/nix/zsh.nix { inherit pkgs; };
  aliases = import ./tools/config/nix/aliases.nix { inherit pkgs; };
  integrations = import ./tools/config/nix/integrations.nix { inherit pkgs; };
  packages = import ./tools/config/nix/packages.nix { inherit pkgs; };

in

pkgs.mkShell {
  # All packages from the packages module
  buildInputs = packages.packages;

  shellHook = ''
    # Set environment variables
    export SHELL=${pkgs.zsh}/bin/zsh
    export DEV_CONFIG=".config"
    export ZDOTDIR="$(pwd)/$DEV_CONFIG/zsh"

    # Create .config/zsh directory in current working directory
    mkdir -p $ZDOTDIR
    
    # Create .zshrc in the .config/zsh directory
    cat > "$ZDOTDIR/.zshrc" << 'EOF'
    # Source your existing zshrc if it exists
    [[ -f ~/.zshrc ]] && source ~/.zshrc

    # Environment setup
    ${environment.setup}

    # ZSH configuration
    ${zshConfig.config}

    # Aliases setup
    ${aliases.setup}

    # Integrations setup
    ${integrations.setup}
EOF
    
    # Launch zsh with ZDOTDIR set
    exec ${pkgs.zsh}/bin/zsh
  '';
}