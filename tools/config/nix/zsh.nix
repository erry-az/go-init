{ pkgs }:
{
  # ZSH-specific configuration that only loads when in zsh
  config = ''
    # Only load zsh-specific configuration if we're actually in zsh
    if [[ -n "$ZSH_VERSION" ]]; then
      # Load zsh plugins directly from nix packages
      source ${pkgs.zsh-fzf-tab}/share/fzf-tab/fzf-tab.plugin.zsh
      source ${pkgs.zsh-fast-syntax-highlighting}/share/zsh/site-functions/fast-syntax-highlighting.plugin.zsh
      source ${pkgs.zsh-autosuggestions}/share/zsh-autosuggestions/zsh-autosuggestions.zsh

      # Add completions to fpath
      fpath=(${pkgs.zsh-completions}/share/zsh/site-functions $fpath)

      # Load completions
      autoload -Uz compinit && compinit

      # Key bindings
      bindkey -e
      bindkey '^p' history-search-backward
      bindkey '^n' history-search-forward
      bindkey '^[w' kill-region

      # History configuration
      HISTSIZE=5000
      HISTFILE=$ZDOTDIR/.zsh_history
      SAVEHIST=$HISTSIZE
      HISTDUP=erase
      setopt appendhistory
      setopt sharehistory
      setopt hist_ignore_space
      setopt hist_ignore_all_dups
      setopt hist_save_no_dups
      setopt hist_ignore_dups
      setopt hist_find_no_dups

      # Completion styling
      zstyle ':completion:*' matcher-list 'm:{a-z}={A-Za-z}'
      zstyle ':completion:*' list-colors "''${(s.:.)LS_COLORS}"
      zstyle ':completion:*' menu no
      zstyle ':fzf-tab:complete:cd:*' fzf-preview 'eza -1 --icons -a --group-directories-first --git --color=always $realpath'
      zstyle ':fzf-tab:complete:__zoxide_z:*' fzf-preview 'eza -1 --icons -a --group-directories-first --git --color=always $realpath'

      # ZSH options
      setopt auto_cd
    fi
  '';
}