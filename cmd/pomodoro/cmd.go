package pomodoro

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

type Cmd struct {
	Status   StatusCmd   `cmd:"" help:"Show Pomodoro status."`
	Settings SettingsCmd `cmd:"" help:"Show Pomodoro settings."`
	Start    StartCmd    `cmd:"" help:"Start a Pomodoro session."`
	Stop     StopCmd     `cmd:"" help:"Stop the current Pomodoro session."`
	Pause    PauseCmd    `cmd:"" help:"Pause the current Pomodoro session."`
}

type (
	StatusCmd   struct{}
	SettingsCmd struct{}
)

type StartCmd struct {
	TimeLeft int `name:"time-left" help:"Remaining time in seconds."`
}

type (
	StopCmd  struct{}
	PauseCmd struct{}
)

const (
	pomodoroStatusQuery = `query {
  pomodoroToday {
    date
    completedCount
    currentRound
    timeLeft
    totalTime
    isRunning
    isPause
    state
  }
}`

	pomodoroSettingsQuery = `query {
  pomodoroSettings {
    workDuration
    shortBreakDuration
    longBreakDuration
    pomodorosBeforeLongBreak
    showNotification
    playSoundOnComplete
  }
}`

	startPomodoroMutation = `mutation startPomodoro($timeLeft: Int!) {
  startPomodoro(timeLeft: $timeLeft)
}`

	stopPomodoroMutation = `mutation {
  stopPomodoro
}`

	pausePomodoroMutation = `mutation {
  pausePomodoro
}`
)

type pomodoroStatusResponse struct {
	Data struct {
		PomodoroToday api.PomodoroToday `json:"pomodoroToday"`
	} `json:"data"`
}

type pomodoroSettingsResponse struct {
	Data struct {
		PomodoroSettings api.PomodoroSettings `json:"pomodoroSettings"`
	} `json:"data"`
}

type pomodoroMutationResponse struct {
	Data struct {
		StartPomodoro bool `json:"startPomodoro"`
		StopPomodoro  bool `json:"stopPomodoro"`
		PausePomodoro bool `json:"pausePomodoro"`
	} `json:"data"`
}

func (c *StatusCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp pomodoroStatusResponse
	if err := apiClient.GraphQL(context.Background(), pomodoroStatusQuery, nil, &resp); err != nil {
		return fmt.Errorf("query pomodoro status: %w", err)
	}

	return printer.Print(resp.Data.PomodoroToday)
}

func (c *SettingsCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp pomodoroSettingsResponse
	if err := apiClient.GraphQL(context.Background(), pomodoroSettingsQuery, nil, &resp); err != nil {
		return fmt.Errorf("query pomodoro settings: %w", err)
	}

	return printer.Print(resp.Data.PomodoroSettings)
}

func (c *StartCmd) Run(apiClient *client.Client, printer output.Printer) error {
	timeLeft := c.TimeLeft
	if timeLeft == 0 {
		var resp pomodoroSettingsResponse
		if err := apiClient.GraphQL(context.Background(), pomodoroSettingsQuery, nil, &resp); err != nil {
			return fmt.Errorf("query pomodoro settings: %w", err)
		}

		timeLeft = resp.Data.PomodoroSettings.WorkDuration
	}

	var resp pomodoroMutationResponse
	if err := apiClient.GraphQL(context.Background(), startPomodoroMutation, map[string]any{
		"timeLeft": timeLeft,
	}, &resp); err != nil {
		return fmt.Errorf("start pomodoro: %w", err)
	}
	if !resp.Data.StartPomodoro {
		return errors.New("start pomodoro: mutation returned false")
	}

	return nil
}

func (c *StopCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp pomodoroMutationResponse
	if err := apiClient.GraphQL(context.Background(), stopPomodoroMutation, nil, &resp); err != nil {
		return fmt.Errorf("stop pomodoro: %w", err)
	}
	if !resp.Data.StopPomodoro {
		return errors.New("stop pomodoro: mutation returned false")
	}

	return nil
}

func (c *PauseCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp pomodoroMutationResponse
	if err := apiClient.GraphQL(context.Background(), pausePomodoroMutation, nil, &resp); err != nil {
		return fmt.Errorf("pause pomodoro: %w", err)
	}
	if !resp.Data.PausePomodoro {
		return errors.New("pause pomodoro: mutation returned false")
	}

	return nil
}
