package videos

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/cmdutil"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	videosQuery = `query videos($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  videos(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    id
    title
    path
    duration
    size
    bucketId
    takenAt
    createdAt
    updatedAt
    tags {
      id
      name
      count
    }
  }
}`

	videoBucketsQuery = `query mediaBuckets($type: DataType!) {
  mediaBuckets(type: $type) {
    id
    name
    itemCount
    topItems
  }
}`

	trashVideosMutation = `mutation trashMediaItems($type: DataType!, $query: String!) {
  trashMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	restoreVideosMutation = `mutation restoreMediaItems($type: DataType!, $query: String!) {
  restoreMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	deleteVideosMutation = `mutation deleteMediaItems($type: DataType!, $query: String!) {
  deleteMediaItems(type: $type, query: $query) {
    type
    query
  }
}`
)

type Cmd struct {
	LS       LSCmd       `cmd:"" help:"List videos."`
	Buckets  BucketsCmd  `cmd:"" help:"List video buckets."`
	Download DownloadCmd `cmd:"" help:"Download a video."`
	Trash    TrashCmd    `cmd:"" help:"Trash videos."`
	Restore  RestoreCmd  `cmd:"" help:"Restore videos from trash."`
	Delete   DeleteCmd   `cmd:"" help:"Delete videos permanently."`
}

type LSCmd struct {
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type BucketsCmd struct{}

type DownloadCmd struct {
	ID  string `arg:"" help:"Video ID."`
	Out string `help:"Local output path. Writes to stdout when omitted."`
}

type TrashCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type RestoreCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type DeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type videosListResponse struct {
	Data struct {
		Videos []api.Video `json:"videos"`
	} `json:"data"`
}

type videoBucketsResponse struct {
	Data struct {
		MediaBuckets []cmdutil.MediaBucket `json:"mediaBuckets"`
	} `json:"data"`
}

type videoMutationResponse struct {
	Data struct {
		TrashMediaItems   cmdutil.MediaMutationResult `json:"trashMediaItems"`
		RestoreMediaItems cmdutil.MediaMutationResult `json:"restoreMediaItems"`
		DeleteMediaItems  cmdutil.MediaMutationResult `json:"deleteMediaItems"`
	} `json:"data"`
}

func (c *LSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	videos, err := listVideos(
		context.Background(),
		apiClient,
		c.Query,
		api.FileSortBy(c.Sort),
		c.Offset,
		c.Limit,
	)
	if err != nil {
		return err
	}

	return printer.PrintList(videos)
}

func (c *BucketsCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp videoBucketsResponse
	if err := apiClient.GraphQL(context.Background(), videoBucketsQuery, map[string]any{
		"type": api.DataTypeVideo.ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("query video buckets: %w", err)
	}

	return printer.PrintList(resp.Data.MediaBuckets)
}

func (c *DownloadCmd) Run(cli *cmdutil.CLIContext, apiClient *client.Client, printer output.Printer) error {
	reader, err := client.DownloadFile(context.Background(), apiClient, "", c.ID)
	if err != nil {
		return fmt.Errorf("download video: %w", err)
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

	if term.IsTerminal(os.Stderr.Fd()) {
		prog := tea.NewProgram(
			vidDownloadProgressModel{bar: progress.New(progress.WithDefaultGradient()), label: c.Out},
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
					prog.Send(vidDownloadProgressMsg(received))
				}
				if readErr != nil {
					break
				}
			}
			prog.Send(vidDownloadDoneMsg{})
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

func (c *TrashCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runVideoMediaMutation(context.Background(), apiClient, trashVideosMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *RestoreCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runVideoMediaMutation(context.Background(), apiClient, restoreVideosMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *DeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runVideoMediaMutation(context.Background(), apiClient, deleteVideosMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func listVideos(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Video, error) {
	if sortBy == "" {
		sortBy = api.FileSortByDateDesc
	}

	if limit > 0 {
		return fetchVideosPage(ctx, apiClient, query, sortBy, offset, limit)
	}

	videos := make([]api.Video, 0, cmdutil.FilesPageSize)
	currentOffset := offset
	for {
		page, err := fetchVideosPage(ctx, apiClient, query, sortBy, currentOffset, cmdutil.FilesPageSize)
		if err != nil {
			return nil, err
		}

		videos = append(videos, page...)
		if len(page) < cmdutil.FilesPageSize {
			return videos, nil
		}

		currentOffset += len(page)
	}
}

func fetchVideosPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Video, error) {
	var resp videosListResponse
	if err := apiClient.GraphQL(ctx, videosQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
		"sortBy": sortBy.ToGraphQL(),
	}, &resp); err != nil {
		return nil, fmt.Errorf("query videos: %w", err)
	}

	return resp.Data.Videos, nil
}

func runVideoMediaMutation(
	ctx context.Context,
	apiClient *client.Client,
	mutation string,
	query string,
) (cmdutil.MediaMutationResult, error) {
	var resp videoMutationResponse
	if err := apiClient.GraphQL(ctx, mutation, map[string]any{
		"query": query,
		"type":  api.DataTypeVideo.ToGraphQL(),
	}, &resp); err != nil {
		return cmdutil.MediaMutationResult{}, fmt.Errorf("run video media mutation: %w", err)
	}

	switch mutation {
	case trashVideosMutation:
		return resp.Data.TrashMediaItems, nil
	case restoreVideosMutation:
		return resp.Data.RestoreMediaItems, nil
	case deleteVideosMutation:
		return resp.Data.DeleteMediaItems, nil
	default:
		return cmdutil.MediaMutationResult{}, errors.New("run video media mutation: unknown mutation")
	}
}

type (
	vidDownloadProgressMsg int64
	vidDownloadDoneMsg     struct{}
)

type vidDownloadProgressModel struct {
	bar      progress.Model
	label    string
	received int64
}

func (m vidDownloadProgressModel) Init() tea.Cmd { return nil }

func (m vidDownloadProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case vidDownloadProgressMsg:
		m.received = int64(v)
		return m, nil
	case vidDownloadDoneMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m vidDownloadProgressModel) View() string {
	const cycleBytes = 1 << 20
	pct := float64(m.received%cycleBytes) / float64(cycleBytes)
	return m.label + " " + m.bar.ViewAs(pct) + fmt.Sprintf(" %d B", m.received) + "\n"
}
