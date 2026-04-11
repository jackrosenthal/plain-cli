package cmd

type VideosCmd struct {
	LS       VideosLSCmd       `cmd:"" help:"List videos."`
	Buckets  VideosBucketsCmd  `cmd:"" help:"List video buckets."`
	Download VideosDownloadCmd `cmd:"" help:"Download a video."`
	Trash    VideosTrashCmd    `cmd:"" help:"Trash videos."`
	Restore  VideosRestoreCmd  `cmd:"" help:"Restore videos from trash."`
	Delete   VideosDeleteCmd   `cmd:"" help:"Delete videos permanently."`
}

type VideosLSCmd struct {
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type VideosBucketsCmd struct{}

type VideosDownloadCmd struct {
	ID  string `arg:"" help:"Video ID."`
	Out string `help:"Local output path. Writes to stdout when omitted."`
}

type VideosTrashCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type VideosRestoreCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type VideosDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}
