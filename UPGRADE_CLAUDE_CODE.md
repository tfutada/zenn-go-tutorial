# Claude Code Upgrade Instructions

## Current Installation
- **Method**: npm (via Volta)
- **Location**: `/Users/tafu/.volta/tools/image/node/22.12.0/bin/claude`
- **Current Version**: 2.0.19

## How to Update

### Check Current Version
```bash
claude --version
```

### Update via npm
```bash
npm update -g @anthropic-ai/claude-code
```

### Check for Updates
```bash
# Check if updates are available
npm outdated -g @anthropic-ai/claude-code

# View latest version on npm registry
npm view @anthropic-ai/claude-code version
```

## Alternative Installation Methods

If you ever need to reinstall or switch methods:

1. **npm** (current method):
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```

2. **Homebrew** (macOS/Linux):
   ```bash
   brew install --cask claude-code
   brew upgrade claude-code  # to update
   ```

3. **macOS/Linux/WSL** (curl):
   ```bash
   curl -fsSL https://claude.ai/install.sh | bash
   ```

4. **Windows PowerShell**:
   ```bash
   irm https://claude.ai/install.ps1 | iex
   ```

## Notes
- Since Claude Code is installed via npm/Volta, always use npm commands for updates
- The update process is safe and preserves your configuration
- Latest stable version as of this documentation: 2.0.19
