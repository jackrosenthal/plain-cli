package screen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

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

	updateScreenMirrorQualityMutation = `mutation updateScreenMirrorQuality($mode: ScreenMirrorMode!) {
  updateScreenMirrorQuality(mode: $mode)
}`
)

type Cmd struct {
	Status  StatusCmd  `cmd:"" help:"Show screen mirror status."`
	Quality QualityCmd `cmd:"" help:"Set screen mirror quality."`
	View    ViewCmd    `cmd:"" help:"View screen mirror in mpv or ffplay."`
}

type StatusCmd struct{}

type QualityCmd struct {
	Mode string `arg:"" help:"Quality mode." enum:"auto,hd,smooth"`
}

type screenStatusResponse struct {
	Data struct {
		ScreenMirrorState          bool                    `json:"screenMirrorState"`
		ScreenMirrorControlEnabled bool                    `json:"screenMirrorControlEnabled"`
		ScreenMirrorQuality        api.ScreenMirrorQuality `json:"screenMirrorQuality"`
	} `json:"data"`
}

type screenMutationResponse struct {
	Data struct {
		StopScreenMirror          bool `json:"stopScreenMirror"`
		UpdateScreenMirrorQuality bool `json:"updateScreenMirrorQuality"`
	} `json:"data"`
}

type screenStatusOutput struct {
	Running        bool                    `json:"running"`
	ControlEnabled bool                    `json:"controlEnabled"`
	Quality        api.ScreenMirrorQuality `json:"quality"`
}

func (c *StatusCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenStatusResponse
	if err := apiClient.GraphQL(context.Background(), screenStatusQuery, nil, &resp); err != nil {
		return fmt.Errorf("query screen mirror status: %w", err)
	}

	return printer.Print(screenStatusOutput{
		Running:        resp.Data.ScreenMirrorState,
		ControlEnabled: resp.Data.ScreenMirrorControlEnabled,
		Quality:        resp.Data.ScreenMirrorQuality,
	})
}

func (c *QualityCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp screenMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateScreenMirrorQualityMutation, map[string]any{
		"mode": api.ScreenMirrorMode(c.Mode).ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("update screen mirror quality: %w", err)
	}
	if !resp.Data.UpdateScreenMirrorQuality {
		return fmt.Errorf("update screen mirror quality: mutation returned false")
	}

	return nil
}

type ViewCmd struct {
	Audio  bool   `help:"Include audio in the stream."`
	Player string `help:"Player to use." enum:"auto,mpv,ffplay" default:"auto"`
}

func (c *ViewCmd) Run(apiClient *client.Client) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	session, track, err := newWebRTCSession(ctx, apiClient, c.Audio)
	if err != nil {
		return err
	}
	defer session.close(apiClient)

	port, sdpPath, sdpCleanup, err := setupRTPForwarding(track.Codec())
	if err != nil {
		return err
	}
	defer sdpCleanup()

	player, playerArgs, err := resolvePlayer(c.Player, sdpPath)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, player, playerArgs...)
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", player, err)
	}

	// Give the player time to bind the UDP port before sending packets.
	time.Sleep(200 * time.Millisecond)
	go forwardRTP(ctx, session.pc, track, fmt.Sprintf("127.0.0.1:%d", port))

	return cmd.Wait()
}
