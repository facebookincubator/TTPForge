# Shell Completion

TTPForge provides shell completion for TTP references and repository names.

## Installation

Generate and source the completion script for your shell:

**Bash:**

```bash
ttpforge completion bash > ~/.ttpforge-completion.bash
echo 'source ~/.ttpforge-completion.bash' >> ~/.bashrc
source ~/.ttpforge-completion.bash
```

**Zsh:**

```zsh
ttpforge completion zsh > ~/.ttpforge-completion.zsh
echo 'source ~/.ttpforge-completion.zsh' >> ~/.zshrc
source ~/.ttpforge-completion.zsh
```

**Fish:**

```fish
ttpforge completion fish > ~/.config/fish/completions/ttpforge.fish
```

**PowerShell:**

```powershell
ttpforge completion powershell > ttpforge.ps1
. .\ttpforge.ps1
```

## Usage

Completion is dynamic and works for both named and positional arguments:

```bash
ttpforge run examples//<TAB>                # Shows TTPs in examples repo
ttpforge enum ttps --repo <TAB>             # Shows: examples, forgearmory, meta-secure
ttpforge --config custom.yaml run <TAB>     # Works with custom configs
```

## Troubleshooting

If completion isn't working:

1. Verify your config: `cat ~/.ttpforge/config.yaml`
2. Test manually: `ttpforge __complete run ""`
3. Ensure repositories are installed: `ttpforge list repos`

Completion automatically falls back to file completion when no TTPs/repos match.
