package cmd

type CallsCmd struct {
	LS     CallsLSCmd     `cmd:"" help:"List calls."`
	Call   CallsCallCmd   `cmd:"" help:"Place a call."`
	Delete CallsDeleteCmd `cmd:"" help:"Delete calls."`
}

type CallsLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type CallsCallCmd struct {
	Number string `arg:"" help:"Phone number to call."`
}

type CallsDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}
