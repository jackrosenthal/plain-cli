package cmd

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

type ScreenCmd struct {
	Status  ScreenStatusCmd  `cmd:"" help:"Show screen mirror status."`
	Start   ScreenStartCmd   `cmd:"" help:"Start screen mirroring."`
	Stop    ScreenStopCmd    `cmd:"" help:"Stop screen mirroring."`
	Quality ScreenQualityCmd `cmd:"" help:"Set screen mirror quality."`
}

type ScreenStatusCmd struct{}

type ScreenStartCmd struct {
	Audio bool `help:"Include audio in the stream."`
}

type ScreenStopCmd struct{}

type ScreenQualityCmd struct {
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

func (c *ScreenStatusCmd) Run(apiClient *client.Client, printer output.Printer) error {
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

func (c *ScreenStartCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenMutationResponse
	if err := apiClient.GraphQL(context.Background(), startScreenMirrorMutation, map[string]any{
		"audio": c.Audio,
	}, &resp); err != nil {
		return fmt.Errorf("start screen mirroring: %w", err)
	}
	if !resp.Data.StartScreenMirror {
		return errors.New("start screen mirroring: mutation returned false")
	}

	message := "Started screen mirroring."
	if c.Audio {
		message = "Started screen mirroring with audio."
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: message,
	})
}

func (c *ScreenStopCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenMutationResponse
	if err := apiClient.GraphQL(context.Background(), stopScreenMirrorMutation, nil, &resp); err != nil {
		return fmt.Errorf("stop screen mirroring: %w", err)
	}
	if !resp.Data.StopScreenMirror {
		return errors.New("stop screen mirroring: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "Stopped screen mirroring.",
	})
}

func (c *ScreenQualityCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateScreenMirrorQualityMutation, map[string]any{
		"mode": api.ScreenMirrorMode(c.Mode).ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("update screen mirror quality: %w", err)
	}
	if !resp.Data.UpdateScreenMirrorQuality {
		return errors.New("update screen mirror quality: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Set screen mirror quality to %s.", c.Mode),
	})
}
