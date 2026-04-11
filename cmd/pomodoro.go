package cmd

type PomodoroCmd struct {
	Status   PomodoroStatusCmd   `cmd:"" help:"Show Pomodoro status."`
	Settings PomodoroSettingsCmd `cmd:"" help:"Show Pomodoro settings."`
	Start    PomodoroStartCmd    `cmd:"" help:"Start a Pomodoro session."`
	Stop     PomodoroStopCmd     `cmd:"" help:"Stop the current Pomodoro session."`
	Pause    PomodoroPauseCmd    `cmd:"" help:"Pause the current Pomodoro session."`
}

type (
	PomodoroStatusCmd   struct{}
	PomodoroSettingsCmd struct{}
)

type PomodoroStartCmd struct {
	TimeLeft int `name:"time-left" help:"Remaining time in seconds."`
}

type (
	PomodoroStopCmd  struct{}
	PomodoroPauseCmd struct{}
)
