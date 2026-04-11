package cmd

type ContactsCmd struct {
	LS      ContactsLSCmd      `cmd:"" help:"List contacts."`
	Sources ContactsSourcesCmd `cmd:"" help:"List contact sources."`
	Delete  ContactsDeleteCmd  `cmd:"" help:"Delete contacts."`
}

type ContactsLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type ContactsSourcesCmd struct{}

type ContactsDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}
