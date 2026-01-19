# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build    # Build binary to bin/jamesbot
make run      # Build and run
make test     # Run tests with race detection: go test -v -race ./...
make fmt      # Format code
make clean    # Remove build artifacts

# Run single test
go test -v -run Test_CommandName ./internal/command/

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Architecture

```
cmd/bot/main.go          Entry point, signal handling, DI wiring
        ↓
internal/config          Viper-based config with JAMESBOT_ env prefix
        ↓
internal/bot             Bot lifecycle (New, Start, Stop, RegisterCommand)
        ↓
internal/handler         Discord event routing
    ├── ready.go         Bot connection events
    └── interaction.go   Slash command dispatch → Registry → Middleware → Execute
        ↓
internal/command         Command framework
    ├── command.go       Command and PermissionedCommand interfaces
    ├── context.go       Execution context with response helpers
    ├── registry.go      Thread-safe command storage (sync.RWMutex)
    └── *.go             Individual command implementations
        ↓
internal/middleware      Composable request pipeline
    ├── middleware.go    Chain() function composes middlewares
    ├── recovery.go      Outermost: catches panics, logs stack traces
    └── logging.go       Inner: logs execution time, user/guild context
        ↓
pkg/errutil              Custom error types with Unwrap() support
```

## Key Patterns

**Command Interface** - All commands implement:
```go
type Command interface {
    Name() string
    Description() string
    Options() []*discordgo.ApplicationCommandOption
    Execute(ctx *Context) error
}
```

For permission-gated commands, also implement `PermissionedCommand` with `Permissions() int64`.

**Middleware Chain** - Executed in order: Recovery → Logging → Command.Execute(). Chain is built via `middleware.Chain(mw1, mw2, ...)`.

**Context Helpers** - Use `ctx.Respond()`, `ctx.RespondEphemeral()`, `ctx.RespondEmbed()` for responses. Use `ctx.StringOption()`, `ctx.UserOption()`, etc. for parameter access.

**Error Handling** - Return `UserFriendlyError{UserMessage: "...", Err: internalErr}` to show user-safe message while logging internal details. InteractionHandler extracts UserMessage for ephemeral responses.

## Adding a New Command

1. Create `internal/command/mycommand.go` implementing `Command` interface
2. Create `internal/command/mycommand_test.go` with table-driven tests
3. Register in `cmd/bot/main.go`: `b.RegisterCommand(&command.MyCommand{})`
4. Run `make test` to verify

## Configuration

Priority: Environment Variables > YAML File > Defaults

| Env Var | Default | Purpose |
|---------|---------|---------|
| `JAMESBOT_DISCORD_TOKEN` | - | Required. Bot token |
| `JAMESBOT_DISCORD_GUILD_ID` | "" | Optional. Set for instant command registration during dev |
| `JAMESBOT_LOGGING_LEVEL` | info | debug, info, warn, error |
| `JAMESBOT_SHUTDOWN_TIMEOUT` | 10s | Graceful shutdown timeout |

## Context Usage

When using the Task tool with specialized agents, prefer it over multiple sequential file reads to save context. The codebase has comprehensive test coverage - run `make test` after changes.
