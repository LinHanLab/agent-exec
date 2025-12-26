# pkg/claude/

Core Claude CLI integration, prompt execution, and streaming JSON parsing.

## Files

- `types.go` - JSON response type definitions for Claude CLI streaming output
- `prompt.go` - Prompt execution (one-shot and iteration loop), signal handling
- `parser.go` - Streaming JSON parser, formats tool usage and results for display
- `validate.go` - Prompt and iteration argument validation

## Conventions

- Validates all inputs before execution
- Handles SIGINT/SIGTERM for graceful shutdown
- Prints formatted output with emojis for different message types
- Streaming JSON parsing with line-by-line processing

---

Update this when: files are added, removed, or change purpose.
