package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jackrosenthal/plain-cli/cmd/audio"
	"github.com/jackrosenthal/plain-cli/cmd/auth"
	"github.com/jackrosenthal/plain-cli/cmd/bookmarks"
	"github.com/jackrosenthal/plain-cli/cmd/calls"
	"github.com/jackrosenthal/plain-cli/cmd/chat"
	"github.com/jackrosenthal/plain-cli/cmd/clipboard"
	"github.com/jackrosenthal/plain-cli/cmd/contacts"
	"github.com/jackrosenthal/plain-cli/cmd/device"
	"github.com/jackrosenthal/plain-cli/cmd/feeds"
	"github.com/jackrosenthal/plain-cli/cmd/files"
	"github.com/jackrosenthal/plain-cli/cmd/images"
	"github.com/jackrosenthal/plain-cli/cmd/notes"
	"github.com/jackrosenthal/plain-cli/cmd/notifications"
	"github.com/jackrosenthal/plain-cli/cmd/packages"
	"github.com/jackrosenthal/plain-cli/cmd/pomodoro"
	"github.com/jackrosenthal/plain-cli/cmd/screen"
	"github.com/jackrosenthal/plain-cli/cmd/sms"
	"github.com/jackrosenthal/plain-cli/cmd/tags"
	"github.com/jackrosenthal/plain-cli/cmd/videos"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/cmdutil"
	"github.com/jackrosenthal/plain-cli/internal/config"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

type CLI struct {
	Host     string        `help:"Plain server base URL." env:"PLAIN_HOST"`
	Token    string        `help:"Auth token (base64)." env:"PLAIN_TOKEN"`
	ClientID string        `name:"client-id" help:"Stable client UUID." env:"PLAIN_CLIENT_ID"`
	Output   output.Format `help:"Output format." env:"PLAIN_OUTPUT" enum:"table,json,plain" default:"table"`

	Auth          auth.Cmd          `cmd:"" help:"Authentication commands."`
	Device        device.Cmd        `cmd:"" help:"Device queries and actions."`
	Files         files.Cmd         `cmd:"" help:"File management commands."`
	Images        images.Cmd        `cmd:"" help:"Image library commands."`
	Videos        videos.Cmd        `cmd:"" help:"Video library commands."`
	Audio         audio.Cmd         `cmd:"" help:"Audio playback and library commands."`
	SMS           sms.Cmd           `cmd:"" help:"SMS and MMS commands."`
	Contacts      contacts.Cmd      `cmd:"" help:"Contact management commands."`
	Calls         calls.Cmd         `cmd:"" help:"Call history and actions."`
	Notes         notes.Cmd         `cmd:"" help:"Note management commands."`
	Feeds         feeds.Cmd         `cmd:"" help:"Feed and feed entry commands."`
	Packages      packages.Cmd      `cmd:"" help:"Package management commands."`
	Notifications notifications.Cmd `cmd:"" help:"Notification commands."`
	Bookmarks     bookmarks.Cmd     `cmd:"" help:"Bookmark commands."`
	Chat          chat.Cmd          `cmd:"" help:"Chat commands."`
	Tags          tags.Cmd          `cmd:"" help:"Tag commands."`
	Screen        screen.Cmd        `cmd:"" help:"Screen mirror commands."`
	Pomodoro      pomodoro.Cmd      `cmd:"" help:"Pomodoro commands."`
	Clipboard     clipboard.Cmd     `cmd:"" help:"Clipboard commands."`
}

func (c *CLI) AfterApply(ctx *kong.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if c.Host == "" {
		c.Host = cfg.Host
	}
	if c.Token == "" {
		c.Token = cfg.Token
	}
	if c.ClientID == "" {
		c.ClientID = cfg.ClientID
	}

	printer := output.New(c.Output, os.Stdout)
	ctx.Bind(&cmdutil.CLIContext{
		Host:     c.Host,
		Token:    c.Token,
		ClientID: c.ClientID,
		Output:   c.Output,
	})
	ctx.BindTo(printer, (*output.Printer)(nil))

	if skipClientForCommand(ctx.Command()) {
		return nil
	}

	apiClient, err := client.NewClient(c.Host, c.ClientID, c.Token)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	if err := apiClient.FetchServerTimeOffset(context.Background()); err != nil {
		return fmt.Errorf("fetch server time offset: %w", err)
	}
	if err := apiClient.FetchURLToken(context.Background()); err != nil {
		return fmt.Errorf("fetch url token: %w", err)
	}

	ctx.Bind(apiClient)

	return nil
}

func skipClientForCommand(command string) bool {
	switch command {
	case "auth login", "auth status":
		return true
	default:
		return false
	}
}
