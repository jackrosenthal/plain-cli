package cmd

type FeedsCmd struct {
	LS      FeedsLSCmd      `cmd:"" help:"List feeds."`
	Add     FeedsAddCmd     `cmd:"" help:"Add a feed."`
	Update  FeedsUpdateCmd  `cmd:"" help:"Update a feed."`
	Delete  FeedsDeleteCmd  `cmd:"" help:"Delete a feed."`
	Sync    FeedsSyncCmd    `cmd:"" help:"Sync feeds."`
	Import  FeedsImportCmd  `cmd:"" help:"Import feeds from OPML."`
	Export  FeedsExportCmd  `cmd:"" help:"Export feeds as OPML."`
	Entries FeedsEntriesCmd `cmd:"" help:"Manage feed entries."`
}

type FeedsLSCmd struct{}

type FeedsAddCmd struct {
	URL          string `arg:"" help:"Feed URL."`
	FetchContent bool   `help:"Fetch article content."`
}

type FeedsUpdateCmd struct {
	ID           string `arg:"" help:"Feed ID."`
	Name         string `help:"Feed name." required:""`
	FetchContent bool   `help:"Fetch article content."`
}

type FeedsDeleteCmd struct {
	ID string `arg:"" help:"Feed ID."`
}

type FeedsSyncCmd struct {
	ID string `help:"Specific feed ID to sync."`
}

type FeedsImportCmd struct {
	OpmlFile string `arg:"" help:"Local OPML file path."`
}

type FeedsExportCmd struct{}

type FeedsEntriesCmd struct {
	LS          FeedsEntriesLSCmd          `cmd:"" help:"List feed entries."`
	Get         FeedsEntriesGetCmd         `cmd:"" help:"Get a feed entry by ID."`
	Delete      FeedsEntriesDeleteCmd      `cmd:"" help:"Delete feed entries."`
	SaveToNotes FeedsEntriesSaveToNotesCmd `cmd:"" help:"Save feed entries to notes."`
}

type FeedsEntriesLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type FeedsEntriesGetCmd struct {
	ID string `arg:"" help:"Feed entry ID."`
}

type FeedsEntriesDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type FeedsEntriesSaveToNotesCmd struct {
	Query string `arg:"" help:"Selection query."`
}
