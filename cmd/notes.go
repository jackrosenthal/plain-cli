package cmd

type NotesCmd struct {
	LS      NotesLSCmd      `cmd:"" help:"List notes."`
	Get     NotesGetCmd     `cmd:"" help:"Get a note by ID."`
	Save    NotesSaveCmd    `cmd:"" help:"Create or update a note."`
	Trash   NotesTrashCmd   `cmd:"" help:"Trash notes."`
	Restore NotesRestoreCmd `cmd:"" help:"Restore notes from trash."`
	Delete  NotesDeleteCmd  `cmd:"" help:"Delete notes permanently."`
	Export  NotesExportCmd  `cmd:"" help:"Export notes."`
}

type NotesLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type NotesGetCmd struct {
	ID string `arg:"" help:"Note ID."`
}

type NotesSaveCmd struct {
	ID      string `help:"Existing note ID."`
	Title   string `help:"Note title." required:""`
	Content string `help:"Note content. Reads stdin when omitted."`
}

type NotesTrashCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type NotesRestoreCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type NotesDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type NotesExportCmd struct {
	Query string `arg:"" help:"Selection query."`
}
