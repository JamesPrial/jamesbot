# JamesBot

A Discord moderation bot built with Go.

## Features

- **Utility Commands**
  - `/ping` - Check if the bot is responsive
  - `/echo` - Echo back a message

- **Moderation Commands**
  - `/kick` - Kick a member from the server
  - `/ban` - Ban a member from the server
  - `/mute` - Timeout a member (1 minute to 28 days)
  - `/warn` - Warn a member via DM

## Setup

### Prerequisites

- Go 1.21 or higher
- A Discord bot token ([Get one here](https://discord.com/developers/applications))

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd jamesbot
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Configure the bot:
   ```bash
   cp config/config.yaml.example config/config.yaml
   # Edit config/config.yaml and add your Discord bot token
   ```

### Configuration

Configuration can be provided via:
- `config/config.yaml` file
- Environment variables with `JAMESBOT_` prefix

#### Config File Example

```yaml
discord:
  token: "YOUR_BOT_TOKEN_HERE"
  guild_id: ""  # Optional: Set for faster command registration during development
  cleanup_on_shutdown: false

logging:
  level: "info"
  format: "console"

shutdown:
  timeout: "10s"
```

#### Environment Variables

```bash
export JAMESBOT_DISCORD_TOKEN="your_token_here"
export JAMESBOT_LOGGING_LEVEL="debug"
export JAMESBOT_SHUTDOWN_TIMEOUT="10s"
```

## Running the Bot

### Using Make

```bash
# Build the bot
make build

# Build and run
make run

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean
```

### Manual Build

```bash
# Build
go build -o bin/jamesbot ./cmd/bot

# Run
./bin/jamesbot
```

### Direct Run

```bash
go run ./cmd/bot
```

## Project Structure

```
jamesbot/
├── cmd/
│   └── bot/
│       └── main.go           # Application entry point
├── internal/
│   ├── bot/
│   │   ├── bot.go           # Core bot implementation
│   │   └── options.go       # Functional options
│   ├── command/
│   │   ├── command.go       # Command interface
│   │   ├── context.go       # Command execution context
│   │   ├── registry.go      # Command registry
│   │   ├── ping.go          # Ping command
│   │   ├── echo.go          # Echo command
│   │   ├── kick.go          # Kick command
│   │   ├── ban.go           # Ban command
│   │   ├── mute.go          # Mute command
│   │   └── warn.go          # Warn command
│   ├── config/
│   │   ├── config.go        # Configuration structs
│   │   └── loader.go        # Configuration loading
│   ├── handler/
│   │   ├── interaction.go   # Interaction event handler
│   │   └── ready.go         # Ready event handler
│   └── middleware/
│       ├── middleware.go    # Middleware interface
│       ├── recovery.go      # Panic recovery middleware
│       └── logging.go       # Logging middleware
├── pkg/
│   └── errutil/
│       └── errors.go        # Custom error types
├── config/
│   └── config.yaml.example  # Example configuration
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── Makefile                 # Build automation
└── README.md               # This file
```

## Development

### Adding New Commands

1. Create a new file in `internal/command/` (e.g., `mycommand.go`)
2. Implement the `Command` interface:
   ```go
   type MyCommand struct{}

   func (c *MyCommand) Name() string {
       return "mycommand"
   }

   func (c *MyCommand) Description() string {
       return "Description of my command"
   }

   func (c *MyCommand) Options() []*discordgo.ApplicationCommandOption {
       // Define command options
       return nil
   }

   func (c *MyCommand) Execute(ctx *Context) error {
       // Implement command logic
       return ctx.Respond("Hello!")
   }
   ```

3. Register the command in `cmd/bot/main.go`:
   ```go
   if err := b.RegisterCommand(&command.MyCommand{}); err != nil {
       logger.Fatal().Err(err).Msg("failed to register mycommand")
   }
   ```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

## License

[Add your license here]
