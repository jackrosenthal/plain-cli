package cmd

type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Authenticate with a Plain device."`
	Status AuthStatusCmd `cmd:"" help:"Check whether the current token is valid."`
}

type AuthLoginCmd struct {
	Password bool `help:"Prompt for a password before authenticating."`
}

type AuthStatusCmd struct{}
