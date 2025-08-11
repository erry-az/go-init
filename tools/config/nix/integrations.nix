{ pkgs }:
{
  # Shell integrations for various tools
  setup = ''
    # Starship prompt - only load if available
    if command -v ${pkgs.starship}/bin/starship >/dev/null 2>&1; then
      if [[ -n "$ZSH_VERSION" ]]; then
        eval "$(${pkgs.starship}/bin/starship init zsh)"
      else
        eval "$(${pkgs.starship}/bin/starship init bash)"
      fi
    fi
    
    # Zoxide directory jumping - only load if available
    if command -v ${pkgs.zoxide}/bin/zoxide >/dev/null 2>&1; then
      if [[ -n "$ZSH_VERSION" ]]; then
        eval "$(${pkgs.zoxide}/bin/zoxide init zsh)"
        eval "$(${pkgs.zoxide}/bin/zoxide init --cmd cd zsh)"
      else
        eval "$(${pkgs.zoxide}/bin/zoxide init bash)"
      fi
    fi
    
    # FZF fuzzy finder - only load if available
    if command -v ${pkgs.fzf}/bin/fzf >/dev/null 2>&1; then
      if [[ -n "$ZSH_VERSION" ]]; then
        eval "$(${pkgs.fzf}/bin/fzf --zsh)"
      else
        eval "$(${pkgs.fzf}/bin/fzf --bash)"
      fi
    fi
  '';
}