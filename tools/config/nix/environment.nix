{ pkgs }:
{
  # Basic environment setup
  setup = ''
    echo "Setting up nix-shell environment..."

    # Set zsh as default shell but don't auto-exec for non-interactive sessions
    export SHELL=${pkgs.zsh}/bin/zsh

    # Set up LS_COLORS if not defined
    if [[ -z "$LS_COLORS" ]]; then
      export LS_COLORS="di=34:ln=35:so=32:pi=33:ex=31:bd=34;46:cd=34;43:su=30;41:sg=30;46:tw=30;42:ow=30;43"
    fi

    # Environment variables
    export EDITOR=nvim

    # Go environment variables
    export GOROOT=${pkgs.go}/share/go
    export GOPATH=$HOME/go
    export GOBIN=$GOPATH/bin
    export PATH=$GOBIN:$PATH

    # FZF theme configuration
    export FZF_DEFAULT_OPTS="
        --color=fg:#908caa,bg:-1,hl:#ebbcba
        --color=fg+:#e0def4,bg+:#26233a,hl+:#ebbcba
        --color=border:#403d52,header:#31748f,gutter:#191724
        --color=spinner:#f6c177,info:#9ccfd8
        --color=pointer:#c4a7e7,marker:#eb6f92,prompt:#908caa"
  '';
}