package cmd

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
