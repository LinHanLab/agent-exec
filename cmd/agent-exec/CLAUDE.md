# cmd/agent-exec/

CLI entry point using Cobra for command parsing.

## Files

- `root.go` - Root command definition and main() entry point
- `run.go` - Run subcommand with -n/--iterations and -s/--sleep flags

## Conventions

- Uses Cobra for command parsing and help generation
- Flags use short (-n, -s) and long (--iterations, --sleep) forms

---

Update this when: files are added, removed, or change purpose.
