package cmd

type ImagesCmd struct {
	LS       ImagesLSCmd       `cmd:"" help:"List images."`
	Buckets  ImagesBucketsCmd  `cmd:"" help:"List image buckets."`
	Download ImagesDownloadCmd `cmd:"" help:"Download an image."`
	Trash    ImagesTrashCmd    `cmd:"" help:"Trash images."`
	Restore  ImagesRestoreCmd  `cmd:"" help:"Restore images from trash."`
	Delete   ImagesDeleteCmd   `cmd:"" help:"Delete images permanently."`
	Search   ImagesSearchCmd   `cmd:"" help:"Manage image search."`
}

type ImagesLSCmd struct {
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type ImagesBucketsCmd struct{}

type ImagesDownloadCmd struct {
	ID  string `arg:"" help:"Image ID."`
	Out string `help:"Local output path. Writes to stdout when omitted."`
}

type ImagesTrashCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type ImagesRestoreCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type ImagesDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type ImagesSearchCmd struct {
	Status  ImagesSearchStatusCmd  `cmd:"" help:"Show image search status."`
	Enable  ImagesSearchEnableCmd  `cmd:"" help:"Enable image search."`
	Disable ImagesSearchDisableCmd `cmd:"" help:"Disable image search."`
	Index   ImagesSearchIndexCmd   `cmd:"" help:"Start indexing images."`
}

type (
	ImagesSearchStatusCmd  struct{}
	ImagesSearchEnableCmd  struct{}
	ImagesSearchDisableCmd struct{}
)

type ImagesSearchIndexCmd struct {
	Force bool `help:"Force a full reindex."`
}
