// JamesBot is a Discord moderation bot built with Go.
package main

import (
	"os"

	"jamesbot/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
