package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/cmdutil"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
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

type Cmd struct {
	LS        LSCmd        `cmd:"" help:"List files."`
	Recent    RecentCmd    `cmd:"" help:"List recent files."`
	Info      InfoCmd      `cmd:"" help:"Show file metadata."`
	Mkdir     MkdirCmd     `cmd:"" help:"Create a directory."`
	Get       GetCmd       `cmd:"" help:"Download a file."`
	Put       PutCmd       `cmd:"" help:"Upload a file."`
	Mv        MvCmd        `cmd:"mv" help:"Move or rename a file or directory."`
	Copy      CopyCmd      `cmd:"" help:"Copy a file or directory."`
	Delete    DeleteCmd    `cmd:"" help:"Delete files or directories."`
	Favorites FavoritesCmd `cmd:"" help:"Manage favorite folders."`
}

type LSCmd struct {
	Root   string `arg:"" help:"Root path to list."`
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field." default:"name" enum:"name,name-desc,size,size-desc,date,date-desc"`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type RecentCmd struct{}

type InfoCmd struct {
	Path string `arg:"" help:"Remote path."`
}

type MkdirCmd struct {
	Path string `arg:"" help:"Directory path to create."`
}

type GetCmd struct {
	RemotePath string `arg:"" help:"Remote path."`
	LocalPath  string `arg:"" optional:"" help:"Local destination path, or - for stdout. Defaults to remote filename in current directory."`
}

type PutCmd struct {
	Source     string `arg:"" help:"Local file path, or - for stdin."`
	RemotePath string `arg:"" help:"Remote destination path."`
}

type CopyCmd struct {
	Src       string `arg:"" help:"Source path."`
	Dst       string `arg:"" help:"Destination path."`
	Overwrite bool   `help:"Overwrite an existing destination."`
}

type MvCmd struct {
	Src               string `arg:"" help:"Source path."`
	Dst               string `arg:"" help:"Destination path."`
	NoClobber         bool   `name:"no-clobber" short:"n" help:"Do not overwrite an existing destination."`
	NoTargetDirectory bool   `name:"no-target-directory" short:"T" help:"Treat the destination as a file path even if it exists as a directory."`
}

type DeleteCmd struct {
	Paths []string `arg:"" help:"Paths to delete."`
}

type FavoritesCmd struct {
	LS     FavoritesLSCmd     `cmd:"" help:"List favorite folders."`
	Add    FavoritesAddCmd    `cmd:"" help:"Add a favorite folder."`
	Remove FavoritesRemoveCmd `cmd:"" help:"Remove a favorite folder."`
	Alias  FavoritesAliasCmd  `cmd:"" help:"Set a favorite folder alias."`
}

type FavoritesLSCmd struct{}

type FavoritesAddCmd struct {
	RootPath string `arg:"" help:"Favorite root path."`
	FullPath string `arg:"" help:"Full favorite path."`
}

type FavoritesRemoveCmd struct {
	FullPath string `arg:"" help:"Full favorite path."`
}

type FavoritesAliasCmd struct {
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

type favoriteFoldersResponse struct {
	Data struct {
		App struct {
			FavoriteFolders []api.FavoriteFolder `json:"favoriteFolders"`
		} `json:"app"`
	} `json:"data"`
}

type favoriteFolderMutationResponse struct {
	Data struct {
		AddFavoriteFolder      []api.FavoriteFolder `json:"addFavoriteFolder"`
		RemoveFavoriteFolder   []api.FavoriteFolder `json:"removeFavoriteFolder"`
		SetFavoriteFolderAlias []api.FavoriteFolder `json:"setFavoriteFolderAlias"`
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

func (c *LSCmd) Run(cli *cmdutil.CLIContext, apiClient *client.Client, printer output.Printer) error {
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

func (c *RecentCmd) Run(cli *cmdutil.CLIContext, apiClient *client.Client, printer output.Printer) error {
	var resp filesRecentResponse
	if err := apiClient.GraphQL(context.Background(), recentFilesQuery, nil, &resp); err != nil {
		return fmt.Errorf("query recent files: %w", err)
	}

	if cli != nil && cli.Output == output.FormatJSON {
		return printer.PrintList(resp.Data.RecentFiles)
	}

	return printLSOutputFull(os.Stdout, resp.Data.RecentFiles, cli == nil || cli.Output == output.FormatTable, true)
}

func (c *InfoCmd) Run(apiClient *client.Client, printer output.Printer) error {
	_, fileName := splitRemotePath(c.Path)

	var resp fileInfoResponse
	if err := apiClient.GraphQL(context.Background(), fileInfoQuery, map[string]any{
		"fileName": fileName,
		"id":       "",
		"path":     path.Clean(c.Path),
	}, &resp); err != nil {
		return fmt.Errorf("query file info: %w", err)
	}

	return printer.Print(resp.Data.FileInfo)
}

func (c *MkdirCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp createDirResponse
	if err := apiClient.GraphQL(context.Background(), createDirMutation, map[string]any{
		"path": c.Path,
	}, &resp); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	return printer.Print(resp.Data.CreateDir)
}

func (c *GetCmd) Run(cli *cmdutil.CLIContext, apiClient *client.Client, printer output.Printer) error {
	reader, err := client.DownloadFile(context.Background(), apiClient, c.RemotePath, "")
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	dest := c.LocalPath
	if dest == "-" {
		_, err = io.Copy(os.Stdout, reader)
		return err
	}
	if dest == "" {
		dest = path.Base(c.RemotePath)
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	if term.IsTerminal(os.Stderr.Fd()) {
		prog := tea.NewProgram(
			downloadProgressModel{bar: progress.New(progress.WithDefaultGradient()), label: dest},
			tea.WithOutput(os.Stderr),
			tea.WithInput(nil),
		)
		copyDone := make(chan error, 1)
		go func() {
			var received int64
			buf := make([]byte, 32*1024)
			var copyErr error
			for {
				n, readErr := reader.Read(buf)
				if n > 0 {
					if _, writeErr := file.Write(buf[:n]); writeErr != nil {
						copyErr = writeErr
						break
					}
					received += int64(n)
					prog.Send(downloadProgressMsg(received))
				}
				if readErr != nil {
					break
				}
			}
			prog.Send(downloadDoneMsg{})
			copyDone <- copyErr
		}()
		if _, err := prog.Run(); err != nil {
			return err
		}
		if err := <-copyDone; err != nil {
			return err
		}
	} else {
		if _, err := io.Copy(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (c *PutCmd) Run(cli *cmdutil.CLIContext, apiClient *client.Client, printer output.Printer) error {
	localPath := c.Source

	if c.Source == "-" {
		tmp, err := os.CreateTemp("", "plain-put-*")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		tmpName := tmp.Name()
		defer func() {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}()

		if _, err := io.Copy(tmp, os.Stdin); err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if err := tmp.Close(); err != nil {
			return err
		}
		localPath = tmpName
	}

	var uploadErr error
	if term.IsTerminal(os.Stderr.Fd()) {
		prog := tea.NewProgram(
			uploadProgressModel{bar: progress.New(progress.WithDefaultGradient()), label: c.RemotePath},
			tea.WithOutput(os.Stderr),
			tea.WithInput(nil),
		)
		uploadDone := make(chan error, 1)
		go func() {
			err := client.Upload(context.Background(), apiClient, localPath, c.RemotePath, func(done, total int64) {
				prog.Send(uploadProgressMsg{done: done, total: total})
			})
			prog.Send(uploadDoneMsg{})
			uploadDone <- err
		}()
		if _, err := prog.Run(); err != nil {
			return err
		}
		uploadErr = <-uploadDone
	} else {
		uploadErr = client.Upload(context.Background(), apiClient, localPath, c.RemotePath, nil)
	}
	if uploadErr != nil {
		return fmt.Errorf("upload: %w", uploadErr)
	}

	return nil
}

func (c *CopyCmd) Run(apiClient *client.Client, printer output.Printer) error {
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

	return nil
}

func (c *MvCmd) Run(apiClient *client.Client, printer output.Printer) error {
	src := path.Clean(c.Src)
	dst, err := resolveMvDestination(context.Background(), apiClient, src, c.Dst, c.NoTargetDirectory)
	if err != nil {
		return err
	}

	if dst == src {
		return errors.New("mv: source and destination are the same path")
	}

	if c.NoClobber {
		existing, err := remoteFileAtPath(context.Background(), apiClient, dst)
		if err != nil {
			return err
		}
		if existing != nil {
			return nil
		}
	}

	var resp boolMutationResponse
	if err := apiClient.GraphQL(context.Background(), moveFileMutation, map[string]any{
		"dst":       dst,
		"overwrite": !c.NoClobber,
		"src":       src,
	}, &resp); err != nil {
		return fmt.Errorf("mv: %w", err)
	}
	if !resp.Data.MoveFile {
		return errors.New("mv: mutation returned false")
	}

	return nil
}

func (c *DeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp boolMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteFilesMutation, map[string]any{
		"paths": c.Paths,
	}, &resp); err != nil {
		return fmt.Errorf("delete files: %w", err)
	}
	if !resp.Data.DeleteFiles {
		return errors.New("delete files: mutation returned false")
	}

	return nil
}

func (c *FavoritesLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFoldersResponse
	if err := apiClient.GraphQL(context.Background(), favoriteFoldersQuery, nil, &resp); err != nil {
		return fmt.Errorf("query favorite folders: %w", err)
	}

	return printer.PrintList(resp.Data.App.FavoriteFolders)
}

func (c *FavoritesAddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFolderMutationResponse
	if err := apiClient.GraphQL(context.Background(), addFavoriteFolderMutation, map[string]any{
		"fullPath": c.FullPath,
		"rootPath": c.RootPath,
	}, &resp); err != nil {
		return fmt.Errorf("add favorite folder: %w", err)
	}

	return printer.PrintList(resp.Data.AddFavoriteFolder)
}

func (c *FavoritesRemoveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFolderMutationResponse
	if err := apiClient.GraphQL(context.Background(), removeFavoriteFolderMutation, map[string]any{
		"fullPath": c.FullPath,
	}, &resp); err != nil {
		return fmt.Errorf("remove favorite folder: %w", err)
	}

	return printer.PrintList(resp.Data.RemoveFavoriteFolder)
}

func (c *FavoritesAliasCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp favoriteFolderMutationResponse
	if err := apiClient.GraphQL(context.Background(), setFavoriteFolderAliasMutation, map[string]any{
		"alias":    c.Alias,
		"fullPath": c.FullPath,
	}, &resp); err != nil {
		return fmt.Errorf("set favorite folder alias: %w", err)
	}

	return printer.PrintList(resp.Data.SetFavoriteFolderAlias)
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

	files := make([]api.File, 0, cmdutil.FilesPageSize)
	currentOffset := offset
	for {
		page, err := fetchFilesPage(ctx, apiClient, root, query, sortBy, currentOffset, cmdutil.FilesPageSize)
		if err != nil {
			return nil, err
		}

		files = append(files, page...)
		if len(page) < cmdutil.FilesPageSize {
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

func splitRemotePath(remotePath string) (string, string) {
	cleaned := path.Clean(remotePath)
	dir := path.Dir(cleaned)
	if dir == "." {
		dir = ""
	}

	return dir, path.Base(cleaned)
}

func resolveMvDestination(ctx context.Context, apiClient *client.Client, src, dst string, noTargetDirectory bool) (string, error) {
	cleanedDst := path.Clean(dst)
	if !noTargetDirectory {
		existing, err := remoteFileAtPath(ctx, apiClient, cleanedDst)
		if err != nil {
			return "", err
		}
		if existing != nil && existing.IsDir {
			return path.Join(cleanedDst, path.Base(src)), nil
		}
		if strings.HasSuffix(dst, "/") {
			return "", fmt.Errorf("mv: destination directory does not exist: %s", dst)
		}
	}

	return cleanedDst, nil
}

func remoteFileAtPath(ctx context.Context, apiClient *client.Client, remotePath string) (*api.File, error) {
	dir, _ := splitRemotePath(remotePath)
	if dir == "" {
		dir = "/"
	}

	files, err := listFiles(ctx, apiClient, dir, "", api.FileSortByName, 0, 0)
	if err != nil {
		return nil, err
	}

	cleaned := path.Clean(remotePath)
	for _, file := range files {
		if path.Clean(file.Path) == cleaned {
			match := file
			return &match, nil
		}
	}

	return nil, nil
}

type (
	downloadProgressMsg int64
	downloadDoneMsg     struct{}
)

type downloadProgressModel struct {
	bar      progress.Model
	label    string
	received int64
}

func (m downloadProgressModel) Init() tea.Cmd { return nil }

func (m downloadProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case downloadProgressMsg:
		m.received = int64(v)
		return m, nil
	case downloadDoneMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m downloadProgressModel) View() string {
	const cycleBytes = 1 << 20 // cycle every MB
	pct := float64(m.received%cycleBytes) / float64(cycleBytes)
	return m.label + " " + m.bar.ViewAs(pct) + fmt.Sprintf(" %d B", m.received) + "\n"
}

type (
	uploadProgressMsg struct{ done, total int64 }
	uploadDoneMsg     struct{}
)

type uploadProgressModel struct {
	bar   progress.Model
	label string
	done  int64
	total int64
}

func (m uploadProgressModel) Init() tea.Cmd { return nil }

func (m uploadProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case uploadProgressMsg:
		m.done = v.done
		m.total = v.total
		return m, nil
	case uploadDoneMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m uploadProgressModel) View() string {
	ratio := 0.0
	if m.total > 0 {
		ratio = float64(m.done) / float64(m.total)
	}
	return m.label + " " + m.bar.ViewAs(ratio) + fmt.Sprintf(" %d/%d B", m.done, m.total) + "\n"
}
