package cmd

type AudioCmd struct {
	LS       AudioLSCmd       `cmd:"" help:"List audio files."`
	Play     AudioPlayCmd     `cmd:"" help:"Play an audio file."`
	Mode     AudioModeCmd     `cmd:"" help:"Set playback mode."`
	Trash    AudioTrashCmd    `cmd:"" help:"Trash audio files."`
	Restore  AudioRestoreCmd  `cmd:"" help:"Restore audio files from trash."`
	Delete   AudioDeleteCmd   `cmd:"" help:"Delete audio files permanently."`
	Playlist AudioPlaylistCmd `cmd:"" help:"Manage the audio playlist."`
}

type AudioLSCmd struct {
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type AudioPlayCmd struct {
	Path string `arg:"" help:"Audio path."`
}

type AudioModeCmd struct {
	Mode string `arg:"" help:"Playback mode." enum:"order,shuffle,repeat,repeat-one"`
}

type AudioTrashCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type AudioRestoreCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type AudioDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type AudioPlaylistCmd struct {
	LS      AudioPlaylistLSCmd      `cmd:"" help:"Show the current playlist."`
	Add     AudioPlaylistAddCmd     `cmd:"" help:"Add audio to the playlist."`
	Remove  AudioPlaylistRemoveCmd  `cmd:"" help:"Remove an item from the playlist."`
	Clear   AudioPlaylistClearCmd   `cmd:"" help:"Clear the playlist."`
	Reorder AudioPlaylistReorderCmd `cmd:"" help:"Reorder the playlist."`
}

type AudioPlaylistLSCmd struct{}

type AudioPlaylistAddCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type AudioPlaylistRemoveCmd struct {
	Path string `arg:"" help:"Audio path to remove."`
}

type AudioPlaylistClearCmd struct{}

type AudioPlaylistReorderCmd struct {
	Paths []string `arg:"" help:"Playlist paths in desired order."`
}
