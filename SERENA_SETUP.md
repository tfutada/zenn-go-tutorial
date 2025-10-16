# Serena MCP Server Setup Guide

## Overview
Serena is an intelligent coding agent MCP server that provides semantic code navigation, analysis, and editing capabilities. It offers symbolic tools for understanding code structure without reading entire files.

## Prerequisites
- Claude Desktop or Claude Code CLI
- Python with `uvx` (from uv package manager)
- Git (for installing from GitHub)

## Installation

### 1. Install uv Package Manager
If you don't have `uv` installed:

```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

### 2. Add Serena MCP Server to Claude

```bash
claude mcp add serena -- uvx --from git+https://github.com/oraios/serena serena start-mcp-server
```

Note: The `--` separator is required to pass the server command with its arguments.

### 3. Verify Installation

```bash
claude mcp list
```

You should see Serena listed with a ✓ Connected status.

## Project Configuration

### 1. Create `.serena` Directory
In your project root, create a `.serena` directory:

```bash
mkdir .serena
```

### 2. Create `project.yml`
Create `.serena/project.yml` with the following configuration:

```yaml
# Language of the project
# Options: csharp, python, rust, java, typescript, go, cpp, ruby
# Note: For C, use cpp. For JavaScript, use typescript
language: typescript  # Change to your project's language

# Use project's gitignore file to ignore files
ignore_all_files_in_gitignore: true

# Additional paths to ignore (same syntax as gitignore)
ignored_paths: []

# Read-only mode (disable all editing tools)
read_only: false

# Tool exclusions (not recommended, but available)
excluded_tools: []

# Initial prompt for the project (optional)
# Shown to LLM upon activating the project
initial_prompt: ""

# Project name (should match your project)
project_name: "your-project-name"
```

### 3. Language-Specific Requirements

#### TypeScript/JavaScript
- No special requirements
- Automatically detects `.ts`, `.tsx`, `.js`, `.jsx` files

#### C#
- Requires a `.sln` file in the project folder

#### Python
- No special requirements
- Automatically detects `.py` files

#### Other Languages
- See Serena documentation for specific requirements

## Available Tools

Serena provides powerful symbolic code navigation and editing tools:

### Code Navigation
- `find_symbol` - Search for symbols by name/path (classes, methods, etc.)
- `get_symbols_overview` - Get high-level overview of file symbols
- `find_referencing_symbols` - Find where a symbol is referenced
- `search_for_pattern` - Flexible regex-based code search

### Code Reading
- `read_file` - Read file content with optional line ranges
- `list_dir` - List files and directories (with recursion)

### Code Editing
- `replace_symbol_body` - Replace entire symbol definition
- `insert_after_symbol` - Insert code after a symbol
- `insert_before_symbol` - Insert code before a symbol
- `replace_regex` - Regex-based code replacement
- `create_text_file` - Create or overwrite files

### Project Management
- `activate_project` - Switch between projects
- `get_current_config` - View current configuration
- `execute_shell_command` - Run shell commands

### Memory System
- `write_memory` - Save project insights for future reference
- `read_memory` - Retrieve saved memories
- `list_memories` - View all memories
- `delete_memory` - Remove a memory

### Thinking Tools
- `think_about_collected_information` - Reflect on gathered data
- `think_about_task_adherence` - Check if on track
- `think_about_whether_you_are_done` - Verify task completion

## Usage Examples

### Example 1: Finding a Class Method
```
Find the getUserProfile method in the User class and show its implementation
```

Serena will use `find_symbol` with `name_path: "User/getUserProfile"` and `include_body: true`.

### Example 2: Understanding File Structure
```
Show me an overview of all symbols in src/services/auth.ts
```

Serena will use `get_symbols_overview` to show classes, functions, and exports.

### Example 3: Finding References
```
Find all places where the calculateDiscount function is called
```

Serena will use `find_referencing_symbols` to locate all usages.

## Best Practices

### 1. Use Symbolic Tools First
Instead of reading entire files, use:
- `get_symbols_overview` for file structure
- `find_symbol` for specific symbols
- Only read full files when necessary

### 2. Leverage the Memory System
Save important project insights:
- Architecture decisions
- Common patterns
- Build/test commands
- Project conventions

### 3. Efficient Code Reading
- Read only necessary code
- Use `depth` parameter in `find_symbol` to get child symbols (e.g., class methods)
- Use `substring_matching` for partial name matches

### 4. Regex-Based Editing
- Use wildcards (.*?) for flexible matching
- Keep regexes concise
- Let Serena handle escaping

## Configuration Tips

### Ignore Files
Add to `ignored_paths` if you want to exclude beyond `.gitignore`:

```yaml
ignored_paths:
  - "**/*.test.ts"
  - "**/node_modules/**"
  - "dist/**"
```

### Read-Only Mode
For code exploration without editing:

```yaml
read_only: true
```

### Multiple Projects
You can configure multiple projects and switch between them:

```bash
# In another project directory
mkdir .serena
# Create project.yml with different project_name
```

Then use `activate_project` tool to switch.

## Troubleshooting

### Server Not Connected
```bash
# Restart MCP servers
claude mcp restart serena
```

### Language Server Issues
If symbols aren't detected properly:
- Verify language setting in `project.yml`
- Check that language-specific requirements are met
- Use `restart_language_server` tool

### Cache Issues
Serena caches symbol information. If you make changes outside Serena:
- Use `restart_language_server` tool
- Or restart the MCP server

## Advanced Features

### Onboarding
Serena can analyze your project structure:
- Use `check_onboarding_performed` to see if onboarding happened
- Use `onboarding` tool to create project understanding

### Modes
Serena supports different operational modes:
- Interactive mode (default)
- Editing mode
- One-shot mode

Switch modes with `switch_modes` tool.

## Resources

- GitHub Repository: https://github.com/oraios/serena
- MCP Documentation: https://modelcontextprotocol.io

## Integration with Claude Code

Serena works seamlessly with Claude Code (CLI) and Claude Desktop:
- Symbolic code navigation reduces token usage
- Memory system provides persistent project context
- Thinking tools ensure thoughtful code changes
- Read-only mode for safe exploration

## Example Workflow

1. **Project Setup**
   ```bash
   mkdir .serena
   # Create project.yml
   ```

2. **Start Coding Session**
   ```
   Show me the project structure
   ```

3. **Navigate Code**
   ```
   Find all API routes in the project
   ```

4. **Make Changes**
   ```
   Add error handling to the createUser function
   ```

5. **Save Insights**
   ```
   Save the API architecture to memory for future reference
   ```

## Notes

- Serena is particularly efficient for large codebases
- Symbolic tools reduce token consumption significantly
- Memory system builds project knowledge over time
- Works with any language supported by LSP (Language Server Protocol)

## Support

For issues or questions:
- Check GitHub issues: https://github.com/oraios/serena/issues
- Review tool descriptions with: `get_current_config`