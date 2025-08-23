{
  description = "Go development environment with modern tooling";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Import modular configuration with pkgs
        environment = import ./tools/config/nix/environment.nix { inherit pkgs; };
        zshConfig = import ./tools/config/nix/zsh.nix { inherit pkgs; };
        aliases = import ./tools/config/nix/aliases.nix { inherit pkgs; };
        integrations = import ./tools/config/nix/integrations.nix { inherit pkgs; };
        packages = import ./tools/config/nix/packages.nix { inherit pkgs; };
        
      in {
        devShells.default = pkgs.mkShell {
          # Use modern packages attribute instead of buildInputs
          packages = packages.packages;

          shellHook = ''
            # Preserve original prompt and environment
            export NIX_SHELL_PRESERVE_PROMPT=1
            
            # Set environment variables
            export SHELL=${pkgs.zsh}/bin/zsh
            export DEV_CONFIG=".config"
            export ZDOTDIR="$(pwd)/$DEV_CONFIG/zsh"

            # Create .config/zsh directory in current working directory
            mkdir -p $ZDOTDIR
            
            # Create .zshrc in the .config/zsh directory
            cat > "$ZDOTDIR/.zshrc" << 'EOF'
# Source your existing zshrc if it exists
# [[ -f ~/.zshrc ]] && source ~/.zshrc

# Environment setup
${environment.setup}

# ZSH configuration
${zshConfig.config}

# Aliases setup
${aliases.setup}

# Integrations setup
${integrations.setup}
EOF
            
            echo "🚀 Go development environment loaded!"
            echo "📁 Platform: ${system}"
            echo "🐹 Go version: $(${pkgs.go_1_24}/bin/go version)"
            echo "🔧 Go root: ${pkgs.go_1_24}/share/go"
            echo "📍 Go binary: $(which go)"
            
            # Only launch zsh if this is an interactive session and not already in zsh
            if [[ $- == *i* ]] && [[ -z "$ZSH_VERSION" ]]; then
              exec ${pkgs.zsh}/bin/zsh
            fi
          '';
          
          # Additional metadata
          meta = {
            description = "Cross-platform Go development environment";
            platforms = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
          };
        };

        # Alias for convenience
        devShell = self.devShells.${system}.default;
      }
    );
}