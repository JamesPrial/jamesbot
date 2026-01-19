# JamesBot

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/JamesPrial/jamesbot)](https://goreportcard.com/report/github.com/JamesPrial/jamesbot)

A Discord moderation bot built with Go, featuring slash commands, middleware architecture, and graceful shutdown.

## Features

### Utility Commands
| Command | Description |
|---------|-------------|
| `/ping` | Check bot latency and responsiveness |
| `/echo` | Echo back a message |

### Moderation Commands
| Command | Description | Required Permission |
|---------|-------------|---------------------|
| `/kick` | Kick a member from the server | Kick Members |
| `/ban` | Ban a member (with optional message deletion) | Ban Members |
| `/mute` | Timeout a member (1 minute to 28 days) | Moderate Members |
| `/warn` | Issue a warning to a member via DM | Moderate Members |

### Architecture Highlights
- **Middleware Pattern**: Composable request handling with logging and panic recovery
- **Graceful Shutdown**: Clean disconnection with configurable timeout
- **Thread-Safe Registry**: Concurrent command registration and lookup
- **Custom Error Types**: Structured error handling with user-friendly messages
- **Structured Logging**: JSON or console output with zerolog

## Quick Start

### Prerequisites
- Go 1.21 or higher
- A Discord bot token ([Discord Developer Portal](https://discord.com/developers/applications))

### Installation

```bash
# Clone the repository
git clone https://github.com/JamesPrial/jamesbot.git
cd jamesbot

# Install dependencies
go mod download

# Build
make build
```

### Configuration

#### Option 1: Environment Variable (Recommended)
```bash
export JAMESBOT_DISCORD_TOKEN="your-bot-token"
./bin/jamesbot
```

#### Option 2: Configuration File
```bash
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml with your token
./bin/jamesbot
```

### Configuration Options

| Variable | Config Key | Default | Description |
|----------|------------|---------|-------------|
| `JAMESBOT_DISCORD_TOKEN` | `discord.token` | - | **Required.** Discord bot token |
| `JAMESBOT_DISCORD_GUILD_ID` | `discord.guild_id` | `""` | Guild ID for faster dev registration |
| `JAMESBOT_LOGGING_LEVEL` | `logging.level` | `info` | Log level (debug, info, warn, error) |
| `JAMESBOT_LOGGING_FORMAT` | `logging.format` | `console` | Log format (console, json) |
| `JAMESBOT_SHUTDOWN_TIMEOUT` | `shutdown.timeout` | `10s` | Graceful shutdown timeout |

## Bot Permissions

When inviting the bot to your server, ensure it has these permissions:

| Permission | Required For |
|------------|--------------|
| Send Messages | Command responses |
| Use Slash Commands | Registering commands |
| Kick Members | `/kick` command |
| Ban Members | `/ban` command |
| Moderate Members | `/mute` and `/warn` commands |

**OAuth2 URL Generator Settings:**
- Scopes: `bot`, `applications.commands`
- Permissions: `Kick Members`, `Ban Members`, `Moderate Members`, `Send Messages`

## Usage

### Make Commands
```bash
make build    # Build the binary to bin/jamesbot
make run      # Build and run the bot
make test     # Run tests with race detection
make fmt      # Format code with gofmt
make clean    # Remove build artifacts
```

### Manual
```bash
# Build and run
go build -o bin/jamesbot ./cmd/bot
./bin/jamesbot

# Or run directly
go run ./cmd/bot
```

## Project Structure

```
jamesbot/
├── cmd/bot/main.go              # Entry point, signal handling
├── internal/
│   ├── bot/                     # Bot lifecycle management
│   │   ├── bot.go               # Start/Stop, command registration
│   │   └── options.go           # Functional options pattern
│   ├── command/                 # Command framework
│   │   ├── command.go           # Command interface
│   │   ├── context.go           # Execution context helpers
│   │   ├── registry.go          # Thread-safe command registry
│   │   ├── ping.go, echo.go     # Utility commands
│   │   └── kick.go, ban.go, mute.go, warn.go  # Moderation
│   ├── config/                  # Configuration
│   │   ├── config.go            # Config structs
│   │   └── loader.go            # Viper-based loading
│   ├── handler/                 # Discord event handlers
│   │   ├── interaction.go       # Slash command routing
│   │   └── ready.go             # Bot ready event
│   └── middleware/              # Request middleware
│       ├── middleware.go        # Chain composition
│       ├── logging.go           # Command logging
│       └── recovery.go          # Panic recovery
├── pkg/errutil/                 # Custom error types
├── config/config.example.yaml   # Example configuration
├── Makefile                     # Build automation
├── CHANGELOG.md                 # Version history
└── LICENSE                      # MIT License
```

## Development

### Adding a New Command

1. Create `internal/command/mycommand.go`:
```go
package command

import "github.com/bwmarrin/discordgo"

type MyCommand struct{}

func (c *MyCommand) Name() string        { return "mycommand" }
func (c *MyCommand) Description() string { return "Does something cool" }
func (c *MyCommand) Options() []*discordgo.ApplicationCommandOption { return nil }

func (c *MyCommand) Execute(ctx *Context) error {
    return ctx.Respond("Hello from my command!")
}
```

2. Register in `cmd/bot/main.go`:
```go
if err := b.RegisterCommand(&command.MyCommand{}); err != nil {
    logger.Fatal().Err(err).Msg("failed to register mycommand")
}
```

### Adding a Permissioned Command

Implement the `PermissionedCommand` interface:
```go
func (c *MyCommand) Permissions() int64 {
    return discordgo.PermissionManageMessages
}
```

### Running Tests

```bash
# All tests with race detection
make test

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

| Package | Coverage |
|---------|----------|
| pkg/errutil | 100% |
| internal/config | 96% |
| internal/middleware | 97% |
| internal/handler | 93% |
| internal/command | 71% |
| internal/bot | 56% |

## Dependencies

| Package | Purpose |
|---------|---------|
| [discordgo](https://github.com/bwmarrin/discordgo) | Discord API client |
| [viper](https://github.com/spf13/viper) | Configuration management |
| [zerolog](https://github.com/rs/zerolog) | Structured logging |
| [testify](https://github.com/stretchr/testify) | Test assertions |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [DiscordGo](https://github.com/bwmarrin/discordgo) for the excellent Discord API library
- The Go community for the amazing tooling and ecosystem
