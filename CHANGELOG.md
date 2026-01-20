# Changelog

All notable changes to JamesBot will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-01-20

### Added
- CLI subcommand system using Go stdlib flag package
- `serve` command - starts Discord bot with control API
- `stats` command - displays bot statistics (human-readable and JSON output)
- `rules` command with `list` and `set` subcommands for rule management
- Control API server on localhost:8765 for CLI-bot communication
- New packages: `internal/cli`, `internal/api`, `internal/control`

### Changed
- Refactored `cmd/bot/main.go` to use CLI dispatcher
- Bot now starts via `jamesbot serve` instead of direct execution

### Technical
- 95% test coverage on API client
- 88% test coverage on CLI framework
- No new external dependencies (stdlib only for CLI)

## [1.0.0] - 2025-01-19

### Added

#### Commands
- `/ping` - Check bot latency and responsiveness
- `/echo` - Echo back user input
- `/kick` - Kick a member from the server (requires Kick Members permission)
- `/ban` - Ban a member with optional message deletion period (requires Ban Members permission)
- `/mute` - Timeout a member for 1 minute to 28 days (requires Moderate Members permission)
- `/warn` - Issue a warning to a member via DM (requires Moderate Members permission)

#### Architecture
- Command interface with `Name()`, `Description()`, `Options()`, and `Execute()` methods
- `PermissionedCommand` interface for commands requiring Discord permissions
- Thread-safe command registry with `Register()`, `Get()`, and `All()` methods
- Command context with helper methods: `Respond()`, `RespondEphemeral()`, `RespondEmbed()`
- Option accessors: `StringOption()`, `IntOption()`, `UserOption()`, `BoolOption()`

#### Middleware
- Middleware chain pattern for composable request handling
- Logging middleware - tracks command name, user ID, guild ID, and execution duration
- Recovery middleware - catches panics, logs stack traces, returns user-friendly errors

#### Configuration
- YAML configuration file support via Viper
- Environment variable overrides with `JAMESBOT_` prefix
- Configurable logging level and format (console/JSON)
- Configurable shutdown timeout

#### Error Handling
- `ConfigError` - configuration validation failures
- `ValidationError` - input validation failures
- `CommandError` - command execution failures with wrapped errors
- `UserFriendlyError` - separates internal errors from user-facing messages
- `PermissionError` - missing permission errors

#### Bot Lifecycle
- Graceful startup with slash command registration
- Signal handling for SIGINT and SIGTERM
- Graceful shutdown with configurable timeout
- Optional command cleanup on shutdown

#### Development
- Comprehensive test suite with table-driven tests
- Race condition testing
- Benchmark tests for performance-critical paths
- Makefile with build, test, run, and clean targets

### Technical Details
- Built with Go 1.21+
- Uses discordgo v0.28+ for Discord API
- Structured logging with zerolog
- Configuration management with Viper
- Test assertions with testify

[1.0.0]: https://github.com/JamesPrial/jamesbot/releases/tag/v1.0.0
