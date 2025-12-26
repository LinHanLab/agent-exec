# cmd/

CLI entry point and command routing for the agent-exec tool.

## Files

- `main.go` - Entry point, parses CLI arguments, routes to one-shot or iterations subcommands

## Conventions

- Command-line argument validation happens before delegating to pkg/claude
- Error messages are printed to stderr
- Exit code 130 used for interrupt signals (SIGINT)

---

Update this when: files are added, removed, or change purpose.
