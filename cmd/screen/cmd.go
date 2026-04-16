package screen

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	screenStatusQuery = `query {
  screenMirrorState
  screenMirrorControlEnabled
  screenMirrorQuality {
    mode
    resolution
  }
}`

	startScreenMirrorMutation = `mutation startScreenMirror($audio: Boolean!) {
  startScreenMirror(audio: $audio)
}`

	stopScreenMirrorMutation = `mutation {
  stopScreenMirror
}`

	updateScreenMirrorQualityMutation = `mutation updateScreenMirrorQuality($mode: ScreenMirrorMode!) {
  updateScreenMirrorQuality(mode: $mode)
}`
)

type Cmd struct {
	Status  StatusCmd  `cmd:"" help:"Show screen mirror status."`
	Start   StartCmd   `cmd:"" help:"Start screen mirroring."`
	Stop    StopCmd    `cmd:"" help:"Stop screen mirroring."`
	Quality QualityCmd `cmd:"" help:"Set screen mirror quality."`
}

type StatusCmd struct{}

type StartCmd struct {
	Audio bool `help:"Include audio in the stream."`
}

type StopCmd struct{}

type QualityCmd struct {
	Mode string `arg:"" help:"Quality mode." enum:"auto,hd,smooth"`
}

type screenStatusResponse struct {
	Data struct {
		ScreenMirrorState          string                  `json:"screenMirrorState"`
		ScreenMirrorControlEnabled bool                    `json:"screenMirrorControlEnabled"`
		ScreenMirrorQuality        api.ScreenMirrorQuality `json:"screenMirrorQuality"`
	} `json:"data"`
}

type screenMutationResponse struct {
	Data struct {
		StartScreenMirror         bool `json:"startScreenMirror"`
		StopScreenMirror          bool `json:"stopScreenMirror"`
		UpdateScreenMirrorQuality bool `json:"updateScreenMirrorQuality"`
	} `json:"data"`
}

type screenStatusOutput struct {
	State          string                  `json:"state"`
	ControlEnabled bool                    `json:"controlEnabled"`
	Quality        api.ScreenMirrorQuality `json:"quality"`
}

func (c *StatusCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenStatusResponse
	if err := apiClient.GraphQL(context.Background(), screenStatusQuery, nil, &resp); err != nil {
		return fmt.Errorf("query screen mirror status: %w", err)
	}

	return printer.Print(screenStatusOutput{
		State:          resp.Data.ScreenMirrorState,
		ControlEnabled: resp.Data.ScreenMirrorControlEnabled,
		Quality:        resp.Data.ScreenMirrorQuality,
	})
}

func (c *StartCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenMutationResponse
	if err := apiClient.GraphQL(context.Background(), startScreenMirrorMutation, map[string]any{
		"audio": c.Audio,
	}, &resp); err != nil {
		return fmt.Errorf("start screen mirroring: %w", err)
	}
	if !resp.Data.StartScreenMirror {
		return errors.New("start screen mirroring: mutation returned false")
	}

	return nil
}

func (c *StopCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenMutationResponse
	if err := apiClient.GraphQL(context.Background(), stopScreenMirrorMutation, nil, &resp); err != nil {
		return fmt.Errorf("stop screen mirroring: %w", err)
	}
	if !resp.Data.StopScreenMirror {
		return errors.New("stop screen mirroring: mutation returned false")
	}

	return nil
}

func (c *QualityCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateScreenMirrorQualityMutation, map[string]any{
		"mode": api.ScreenMirrorMode(c.Mode).ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("update screen mirror quality: %w", err)
	}
	if !resp.Data.UpdateScreenMirrorQuality {
		return errors.New("update screen mirror quality: mutation returned false")
	}

	return nil
}
