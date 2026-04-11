package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/config"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

type CLI struct {
	Host     string        `help:"Plain server base URL." env:"PLAIN_HOST"`
	Token    string        `help:"Auth token (base64)." env:"PLAIN_TOKEN"`
	ClientID string        `name:"client-id" help:"Stable client UUID." env:"PLAIN_CLIENT_ID"`
	Output   output.Format `help:"Output format." env:"PLAIN_OUTPUT" enum:"table,json,plain" default:"table"`

	Auth          AuthCmd          `cmd:"" help:"Authentication commands."`
	Device        DeviceCmd        `cmd:"" help:"Device queries and actions."`
	Files         FilesCmd         `cmd:"" help:"File management commands."`
	Images        ImagesCmd        `cmd:"" help:"Image library commands."`
	Videos        VideosCmd        `cmd:"" help:"Video library commands."`
	Audio         AudioCmd         `cmd:"" help:"Audio playback and library commands."`
	SMS           SMSCmd           `cmd:"" help:"SMS and MMS commands."`
	Contacts      ContactsCmd      `cmd:"" help:"Contact management commands."`
	Calls         CallsCmd         `cmd:"" help:"Call history and actions."`
	Notes         NotesCmd         `cmd:"" help:"Note management commands."`
	Feeds         FeedsCmd         `cmd:"" help:"Feed and feed entry commands."`
	Packages      PackagesCmd      `cmd:"" help:"Package management commands."`
	Notifications NotificationsCmd `cmd:"" help:"Notification commands."`
	Bookmarks     BookmarksCmd     `cmd:"" help:"Bookmark commands."`
	Chat          ChatCmd          `cmd:"" help:"Chat commands."`
	Tags          TagsCmd          `cmd:"" help:"Tag commands."`
	Screen        ScreenCmd        `cmd:"" help:"Screen mirror commands."`
	Pomodoro      PomodoroCmd      `cmd:"" help:"Pomodoro commands."`
	Clipboard     ClipboardCmd     `cmd:"" help:"Clipboard commands."`
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
	ctx.Bind(c)
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
