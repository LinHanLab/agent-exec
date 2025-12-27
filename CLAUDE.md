# Documentation Method

This project uses hierarchical CLAUDE.md files.

## Structure

- **Root CLAUDE.md** (this file) - architecture overview, global rules.
- **Folder CLAUDE.md** - each subfolder may have its own CLAUDE.md with purpose, file list, local conventions.

## What to Document

Document what code cannot express: purpose, intent, architectural role.

Avoid documenting what the code already expresses (dependencies, exports, types). This information is discoverable through the LSP and search tools.

## Template

```markdown
# folder-name/

[1-2 line purpose]

## Files

- `file.go` - one-line description
- `file.go` - one-line description

## Conventions

- [Folder-specific rules, if any]

---

Update this when: files are added, removed, or change purpose.
```

# Global Rules

## Checklist Before Commit

Before creating a commit, run both Makefile targets:

```bash
make compile
make quality
```

**Why:** `compile` ensures the code builds successfully. `quality` runs tests, formatting, and linting to catch issues early.

# Architecture Overview

CLI tool wrapping the Claude CLI to run prompts with formatted output. Provides one-shot prompt execution and iterative batch processing with configurable sleep intervals.

## Components

- **cmd/** - CLI entry point and command routing
- **pkg/claude/** - Claude CLI integration, streaming JSON parsing, and prompt execution
- **pkg/format/** - Text formatting utilities for display

---

Update this when: files are added, removed, or change purpose.
