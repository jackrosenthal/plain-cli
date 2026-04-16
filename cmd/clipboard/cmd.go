package clipboard

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const setClipMutation = `mutation setClip($text: String!) {
  setClip(text: $text)
}`

type Cmd struct {
	Set SetCmd `cmd:"" help:"Set the clipboard text."`
}

type SetCmd struct {
	Text string `arg:"" help:"Clipboard text."`
}

type clipboardMutationResponse struct {
	Data struct {
		SetClip bool `json:"setClip"`
	} `json:"data"`
}

func (c *SetCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp clipboardMutationResponse
	if err := apiClient.GraphQL(context.Background(), setClipMutation, map[string]any{
		"text": c.Text,
	}, &resp); err != nil {
		return fmt.Errorf("set clipboard: %w", err)
	}
	if !resp.Data.SetClip {
		return errors.New("set clipboard: mutation returned false")
	}

	return nil
}
