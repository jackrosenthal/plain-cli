package cmd

type FilesCmd struct {
	LS        FilesLSCmd        `cmd:"" help:"List files."`
	Recent    FilesRecentCmd    `cmd:"" help:"List recent files."`
	Info      FilesInfoCmd      `cmd:"" help:"Show file metadata."`
	Mkdir     FilesMkdirCmd     `cmd:"" help:"Create a directory."`
	Write     FilesWriteCmd     `cmd:"" help:"Write a text file."`
	Rename    FilesRenameCmd    `cmd:"" help:"Rename a file or directory."`
	Copy      FilesCopyCmd      `cmd:"" help:"Copy a file or directory."`
	Move      FilesMoveCmd      `cmd:"" help:"Move a file or directory."`
	Delete    FilesDeleteCmd    `cmd:"" help:"Delete files or directories."`
	Download  FilesDownloadCmd  `cmd:"" help:"Download a file."`
	Upload    FilesUploadCmd    `cmd:"" help:"Upload a file."`
	Favorites FilesFavoritesCmd `cmd:"" help:"Manage favorite folders."`
}

type FilesLSCmd struct {
	Root   string `help:"Root path to list."`
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type FilesRecentCmd struct{}

type FilesInfoCmd struct {
	Path string `arg:"" help:"Remote path."`
}

type FilesMkdirCmd struct {
	Path string `arg:"" help:"Directory path to create."`
}

type FilesWriteCmd struct {
	Path      string `arg:"" help:"Remote path."`
	Content   string `help:"File contents. Reads stdin when omitted."`
	Overwrite bool   `help:"Overwrite an existing file."`
}

type FilesRenameCmd struct {
	Path string `arg:"" help:"Existing remote path."`
	Name string `arg:"" help:"New name."`
}

type FilesCopyCmd struct {
	Src       string `arg:"" help:"Source path."`
	Dst       string `arg:"" help:"Destination path."`
	Overwrite bool   `help:"Overwrite an existing destination."`
}

type FilesMoveCmd struct {
	Src       string `arg:"" help:"Source path."`
	Dst       string `arg:"" help:"Destination path."`
	Overwrite bool   `help:"Overwrite an existing destination."`
}

type FilesDeleteCmd struct {
	Paths []string `arg:"" help:"Paths to delete."`
}

type FilesDownloadCmd struct {
	Path string `arg:"" help:"Remote path."`
	Out  string `help:"Local output path. Writes to stdout when omitted."`
}

type FilesUploadCmd struct {
	LocalPath  string `arg:"" help:"Local file path."`
	RemotePath string `arg:"" help:"Remote destination path."`
}

type FilesFavoritesCmd struct {
	LS     FilesFavoritesLSCmd     `cmd:"" help:"List favorite folders."`
	Add    FilesFavoritesAddCmd    `cmd:"" help:"Add a favorite folder."`
	Remove FilesFavoritesRemoveCmd `cmd:"" help:"Remove a favorite folder."`
	Alias  FilesFavoritesAliasCmd  `cmd:"" help:"Set a favorite folder alias."`
}

type FilesFavoritesLSCmd struct{}

type FilesFavoritesAddCmd struct {
	RootPath string `arg:"" help:"Favorite root path."`
	FullPath string `arg:"" help:"Full favorite path."`
}

type FilesFavoritesRemoveCmd struct {
	FullPath string `arg:"" help:"Full favorite path."`
}

type FilesFavoritesAliasCmd struct {
	FullPath string `arg:"" help:"Full favorite path."`
	Alias    string `arg:"" help:"Alias to assign."`
}
