package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	filesPageSize = 100

	filesQuery = `query files($root: String!, $offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  files(root: $root, offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    path
    isDir
    size
    createdAt
    updatedAt
    children
    mediaId
  }
}`

	recentFilesQuery = `query {
  recentFiles {
    path
    isDir
    size
    createdAt
    updatedAt
    children
    mediaId
  }
}`

	fileInfoQuery = `query fileInfo($id: ID!, $path: String!, $fileName: String!) {
  fileInfo(id: $id, path: $path, fileName: $fileName) {
    ... on FileInfo {
      path
      updatedAt
      size
      tags {
        id
        name
      }
    }
    data {
      ... on ImageFileInfo {
        width
        height
        location {
          latitude
          longitude
        }
      }
      ... on VideoFileInfo {
        duration
        width
        height
        location {
          latitude
          longitude
        }
      }
      ... on AudioFileInfo {
        duration
        location {
          latitude
          longitude
        }
      }
    }
  }
}`

	favoriteFoldersQuery = `query {
  app {
    favoriteFolders {
      rootPath
      fullPath
      alias
    }
  }
}`

	createDirMutation = `mutation createDir($path: String!) {
  createDir(path: $path) {
    path
    isDir
    size
    createdAt
    updatedAt
    children
    mediaId
  }
}`

	writeTextFileMutation = `mutation writeTextFile($path: String!, $content: String!, $overwrite: Boolean!) {
  writeTextFile(path: $path, content: $content, overwrite: $overwrite) {
    path
    isDir
    size
    createdAt
    updatedAt
    children
    mediaId
  }
}`

	renameFileMutation = `mutation renameFile($path: String!, $name: String!) {
  renameFile(path: $path, name: $name)
}`

	copyFileMutation = `mutation copyFile($src: String!, $dst: String!, $overwrite: Boolean!) {
  copyFile(src: $src, dst: $dst, overwrite: $overwrite)
}`

	moveFileMutation = `mutation moveFile($src: String!, $dst: String!, $overwrite: Boolean!) {
  moveFile(src: $src, dst: $dst, overwrite: $overwrite)
}`

	deleteFilesMutation = `mutation deleteFiles($paths: [String!]!) {
  deleteFiles(paths: $paths)
}`

	addFavoriteFolderMutation = `mutation addFavoriteFolder($rootPath: String!, $fullPath: String!) {
  addFavoriteFolder(rootPath: $rootPath, fullPath: $fullPath) {
    rootPath
    fullPath
    alias
  }
}`

	removeFavoriteFolderMutation = `mutation removeFavoriteFolder($fullPath: String!) {
  removeFavoriteFolder(fullPath: $fullPath) {
    rootPath
    fullPath
    alias
  }
}`

	setFavoriteFolderAliasMutation = `mutation setFavoriteFolderAlias($fullPath: String!, $alias: String!) {
  setFavoriteFolderAlias(fullPath: $fullPath, alias: $alias) {
    rootPath
    fullPath
    alias
  }
}`
)

var progressBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

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
	Root   string `arg:"" help:"Root path to list."`
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field." default:"name" enum:"name,name-desc,size,size-desc,date,date-desc"`
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

type filesListResponse struct {
	Data struct {
		Files []api.File `json:"files"`
	} `json:"data"`
}

type filesRecentResponse struct {
	Data struct {
		RecentFiles []api.File `json:"recentFiles"`
	} `json:"data"`
}

type fileInfoResponse struct {
	Data struct {
		FileInfo fileInfo `json:"fileInfo"`
	} `json:"data"`
}

type createDirResponse struct {
	Data struct {
		CreateDir api.File `json:"createDir"`
	} `json:"data"`
}

type writeTextFileResponse struct {
	Data struct {
		WriteTextFile api.File `json:"writeTextFile"`
	} `json:"data"`
}

type favoriteFoldersResponse struct {
	Data struct {
		App struct {
			FavoriteFolders []api.FavoriteFolder `json:"favoriteFolders"`
		} `json:"app"`
	} `json:"data"`
}

type favoriteFolderMutationResponse struct {
	Data struct {
		AddFavoriteFolder      api.FavoriteFolder `json:"addFavoriteFolder"`
		RemoveFavoriteFolder   api.FavoriteFolder `json:"removeFavoriteFolder"`
		SetFavoriteFolderAlias api.FavoriteFolder `json:"setFavoriteFolderAlias"`
	} `json:"data"`
}

type boolMutationResponse struct {
	Data struct {
		RenameFile  bool `json:"renameFile"`
		CopyFile    bool `json:"copyFile"`
		MoveFile    bool `json:"moveFile"`
		DeleteFiles bool `json:"deleteFiles"`
	} `json:"data"`
}

type fileInfo struct {
	Path      string      `json:"path"`
	UpdatedAt string      `json:"updatedAt"`
	Size      int         `json:"size"`
	Tags      []api.Tag   `json:"tags"`
	Data      interface{} `json:"data"`
}

type mutationStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (c *FilesLSCmd) Run(cli *CLI, apiClient *client.Client, printer output.Printer) error {
	sortBy := api.FileSortBy(c.Sort)
	files, err := listFiles(context.Background(), apiClient, c.Root, c.Query, sortBy, c.Offset, c.Limit)
	if err != nil {
		return err
	}

	if cli != nil && cli.Output == output.FormatJSON {
		return printer.PrintList(files)
	}

	return printLSOutput(os.Stdout, files, cli == nil || cli.Output == output.FormatTable)
}

var (
	lsDirStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	lsFileStyle = lipgloss.NewStyle()
)

func printLSOutput(w io.Writer, files []api.File, columns bool) error {
	return printLSOutputFull(w, files, columns, false)
}

func printLSOutputFull(w io.Writer, files []api.File, columns bool, fullPath bool) error {
	type entry struct {
		name    string
		display string
		width   int
	}

	entries := make([]entry, 0, len(files))
	maxWidth := 0
	for _, f := range files {
		var name string
		if fullPath {
			name = f.Path
		} else {
			name = path.Base(f.Path)
		}
		var display string
		if f.IsDir {
			display = lsDirStyle.Render(name + "/")
		} else {
			display = lsFileStyle.Render(name)
		}
		entries = append(entries, entry{name: name, display: display, width: lipgloss.Width(name) + boolInt(f.IsDir)})
		if entries[len(entries)-1].width > maxWidth {
			maxWidth = entries[len(entries)-1].width
		}
	}

	if len(entries) == 0 {
		return nil
	}

	if !columns {
		for _, e := range entries {
			if _, err := fmt.Fprintln(w, e.display); err != nil {
				return err
			}
		}
		return nil
	}

	termWidth := 80
	if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil && width > 0 {
		termWidth = width
	}

	colWidth := maxWidth + 2
	numCols := termWidth / colWidth
	if numCols < 1 {
		numCols = 1
	}
	numRows := (len(entries) + numCols - 1) / numCols

	var sb strings.Builder
	for row := range numRows {
		for col := range numCols {
			idx := col*numRows + row
			if idx >= len(entries) {
				break
			}
			e := entries[idx]
			sb.WriteString(e.display)
			// pad to colWidth unless last in row
			isLast := col == numCols-1 || (col+1)*numRows+row >= len(entries)
			if !isLast {
				padding := colWidth - e.width
				sb.WriteString(strings.Repeat(" ", padding))
			}
		}
		sb.WriteByte('\n')
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (c *FilesRecentCmd) Run(cli *CLI, apiClient *client.Client, printer output.Printer) error {
	var resp filesRecentResponse
	if err := apiClient.GraphQL(context.Background(), recentFilesQuery, nil, &resp); err != nil {
		return fmt.Errorf("query recent files: %w", err)
	}

	if cli != nil && cli.Output == output.FormatJSON {
		return printer.PrintList(resp.Data.RecentFiles)
	}

	return printLSOutputFull(os.Stdout, resp.Data.RecentFiles, cli == nil || cli.Output == output.FormatTable, true)
}

func (c *FilesInfoCmd) Run(apiClient *client.Client, printer output.Printer) error {
	dir, fileName := splitRemotePath(c.Path)

	var resp fileInfoResponse
	if err := apiClient.GraphQL(context.Background(), fileInfoQuery, map[string]any{
		"fileName": fileName,
		"id":       "",
		"path":     dir,
	}, &resp); err != nil {
		return fmt.Errorf("query file info: %w", err)
	}

	return printer.Print(resp.Data.FileInfo)
}

func (c *FilesMkdirCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp createDirResponse
	if err := apiClient.GraphQL(context.Background(), createDirMutation, map[string]any{
		"path": c.Path,
	}, &resp); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	return printer.Print(resp.Data.CreateDir)
}

func (c *FilesWriteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	content, err := resolveWriteContent(c.Content)
	if err != nil {
		return err
	}

	var resp writeTextFileResponse
	if err := apiClient.GraphQL(context.Background(), writeTextFileMutation, map[string]any{
		"content":   content,
		"overwrite": c.Overwrite,
		"path":      c.Path,
	}, &resp); err != nil {
		return fmt.Errorf("write text file: %w", err)
	}

	return printer.Print(resp.Data.WriteTextFile)
}

func (c *FilesRenameCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp boolMutationResponse
	if err := apiClient.GraphQL(context.Background(), renameFileMutation, map[string]any{
		"name": c.Name,
		"path": c.Path,
	}, &resp); err != nil {
		return fmt.Errorf("rename file: %w", err)
	}
	if !resp.Data.RenameFile {
		return errors.New("rename file: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "File renamed.",
	})
}

func (c *FilesCopyCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp boolMutationResponse
	if err := apiClient.GraphQL(context.Background(), copyFileMutation, map[string]any{
		"dst":       c.Dst,
		"overwrite": c.Overwrite,
		"src":       c.Src,
	}, &resp); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	if !resp.Data.CopyFile {
		return errors.New("copy file: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "File copied.",
	})
}

func (c *FilesMoveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp boolMutationResponse
	if err := apiClient.GraphQL(context.Background(), moveFileMutation, map[string]any{
		"dst":       c.Dst,
		"overwrite": c.Overwrite,
		"src":       c.Src,
	}, &resp); err != nil {
		return fmt.Errorf("move file: %w", err)
	}
	if !resp.Data.MoveFile {
		return errors.New("move file: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "File moved.",
	})
}

func (c *FilesDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp boolMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteFilesMutation, map[string]any{
		"paths": c.Paths,
	}, &resp); err != nil {
		return fmt.Errorf("delete files: %w", err)
	}
	if !resp.Data.DeleteFiles {
		return errors.New("delete files: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Deleted %d path(s).", len(c.Paths)),
	})
}

func (c *FilesDownloadCmd) Run(cli *CLI, apiClient *client.Client, printer output.Printer) error {
	reader, err := client.DownloadFile(context.Background(), apiClient, c.Path, "")
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	if c.Out == "" {
		_, err = io.Copy(os.Stdout, reader)
		return err
	}

	file, err := os.Create(c.Out)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	writer := io.Writer(file)
	if shouldShowTransferProgress(cli) {
		writer = &progressWriter{
			label: c.Out,
			w:     file,
		}
	}

	if _, err := io.Copy(writer, reader); err != nil {
		return err
	}

	if pw, ok := writer.(*progressWriter); ok {
		pw.Finish()
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Downloaded to %s.", c.Out),
	})
}

func (c *FilesUploadCmd) Run(cli *CLI, apiClient *client.Client, printer output.Printer) error {
	var progress func(done, total int64)
	if shouldShowTransferProgress(cli) {
		renderer := &uploadProgressRenderer{label: c.RemotePath}
		progress = renderer.Update
		defer renderer.Finish()
	}

	if err := client.Upload(context.Background(), apiClient, c.LocalPath, c.RemotePath, progress); err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Uploaded %s to %s.", c.LocalPath, c.RemotePath),
	})
}

func (c *FilesFavoritesLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFoldersResponse
	if err := apiClient.GraphQL(context.Background(), favoriteFoldersQuery, nil, &resp); err != nil {
		return fmt.Errorf("query favorite folders: %w", err)
	}

	return printer.PrintList(resp.Data.App.FavoriteFolders)
}

func (c *FilesFavoritesAddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFolderMutationResponse
	if err := apiClient.GraphQL(context.Background(), addFavoriteFolderMutation, map[string]any{
		"fullPath": c.FullPath,
		"rootPath": c.RootPath,
	}, &resp); err != nil {
		return fmt.Errorf("add favorite folder: %w", err)
	}

	return printer.Print(resp.Data.AddFavoriteFolder)
}

func (c *FilesFavoritesRemoveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFolderMutationResponse
	if err := apiClient.GraphQL(context.Background(), removeFavoriteFolderMutation, map[string]any{
		"fullPath": c.FullPath,
	}, &resp); err != nil {
		return fmt.Errorf("remove favorite folder: %w", err)
	}

	return printer.Print(resp.Data.RemoveFavoriteFolder)
}

func (c *FilesFavoritesAliasCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFolderMutationResponse
	if err := apiClient.GraphQL(context.Background(), setFavoriteFolderAliasMutation, map[string]any{
		"alias":    c.Alias,
		"fullPath": c.FullPath,
	}, &resp); err != nil {
		return fmt.Errorf("set favorite folder alias: %w", err)
	}

	return printer.Print(resp.Data.SetFavoriteFolderAlias)
}

func listFiles(
	ctx context.Context,
	apiClient *client.Client,
	root string,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.File, error) {
	if root == "" {
		root = "/"
	}
	if sortBy == "" {
		sortBy = api.FileSortByName
	}

	if limit > 0 {
		return fetchFilesPage(ctx, apiClient, root, query, sortBy, offset, limit)
	}

	files := make([]api.File, 0, filesPageSize)
	currentOffset := offset
	for {
		page, err := fetchFilesPage(ctx, apiClient, root, query, sortBy, currentOffset, filesPageSize)
		if err != nil {
			return nil, err
		}

		files = append(files, page...)
		if len(page) < filesPageSize {
			return files, nil
		}

		currentOffset += len(page)
	}
}

func fetchFilesPage(
	ctx context.Context,
	apiClient *client.Client,
	root string,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.File, error) {
	var resp filesListResponse
	if err := apiClient.GraphQL(ctx, filesQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
		"root":   root,
		"sortBy": sortBy.ToGraphQL(),
	}, &resp); err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}

	return resp.Data.Files, nil
}

func resolveWriteContent(content string) (string, error) {
	if content != "" {
		return content, nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	return string(data), nil
}

func splitRemotePath(remotePath string) (string, string) {
	cleaned := path.Clean(remotePath)
	dir := path.Dir(cleaned)
	if dir == "." {
		dir = ""
	}

	return dir, path.Base(cleaned)
}

func shouldShowTransferProgress(cli *CLI) bool {
	if cli == nil {
		return true
	}

	return cli.Output == output.FormatTable
}

type progressWriter struct {
	label string
	w     io.Writer
	done  int64
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.done += int64(n)
	w.render()
	return n, err
}

func (w *progressWriter) Finish() {
	w.render()
	_, _ = fmt.Fprintln(os.Stderr)
}

func (w *progressWriter) render() {
	const width = 24

	progress := int(w.done % int64(width+1))
	bar := strings.Repeat("=", progress)
	if progress < width {
		bar += ">"
		bar += strings.Repeat(" ", width-progress-1)
	}

	line := fmt.Sprintf("\r%s %s %d B", truncateLabel(w.label), progressBarStyle.Render("["+bar+"]"), w.done)
	_, _ = fmt.Fprint(os.Stderr, line)
}

type uploadProgressRenderer struct {
	label string
}

func (r *uploadProgressRenderer) Update(done, total int64) {
	const width = 24

	ratio := 0.0
	if total > 0 {
		ratio = float64(done) / float64(total)
		if ratio > 1 {
			ratio = 1
		}
	}

	filled := int(ratio * width)
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("=", filled) + strings.Repeat(" ", width-filled)
	line := fmt.Sprintf(
		"\r%s %s %3.0f%% (%d/%d B)",
		truncateLabel(r.label),
		progressBarStyle.Render("["+bar+"]"),
		ratio*100,
		done,
		total,
	)
	_, _ = fmt.Fprint(os.Stderr, line)
}

func (r *uploadProgressRenderer) Finish() {
	_, _ = fmt.Fprintln(os.Stderr)
}

func truncateLabel(label string) string {
	const maxLen = 32

	label = strings.TrimSpace(label)
	if len(label) <= maxLen {
		return label
	}

	return "..." + label[len(label)-maxLen+3:]
}
