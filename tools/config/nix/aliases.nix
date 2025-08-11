{ pkgs }:
{
  # Shell aliases that work across bash and zsh
  setup = ''
    # Editor aliases
    # alias v="nvim"
    # alias vim="nvim"
    # alias lkjh="nvim"

    # System aliases
    alias c="clear"
    alias x="exit"

    # Enhanced ls with eza
    alias l="${pkgs.eza}/bin/eza -lh --icons=auto --color=always"
    alias ls="${pkgs.eza}/bin/eza --icons=auto --color=always"
    alias ll="${pkgs.eza}/bin/eza -lha --icons=auto --sort=name --group-directories-first --color=always"
    alias ld="${pkgs.eza}/bin/eza -lhD --icons=auto --color=always"
    alias lt="${pkgs.eza}/bin/eza --icons=auto --tree --color=always"

    # Navigation shortcuts
    alias ...="cd ../.."
    alias .3="cd ../../.."
    alias .4="cd ../../../.."
    alias .5="cd ../../../../.."

    # Directory shortcuts
    alias gr="/"

    # Git tools
    alias lg="lazygit"
  '';
}