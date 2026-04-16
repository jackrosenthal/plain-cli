package images

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
	imagesQuery = `query images($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  images(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    id
    title
    path
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

	imageBucketsQuery = `query mediaBuckets($type: DataType!) {
  mediaBuckets(type: $type) {
    id
    name
    itemCount
    topItems
  }
}`

	imageSearchStatusQuery = `query {
  imageSearchStatus {
    status
    downloadProgress
    errorMessage
    modelSize
    modelDir
    isIndexing
    totalImages
    indexedImages
  }
}`

	trashImagesMutation = `mutation trashMediaItems($type: DataType!, $query: String!) {
  trashMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	restoreImagesMutation = `mutation restoreMediaItems($type: DataType!, $query: String!) {
  restoreMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	deleteImagesMutation = `mutation deleteMediaItems($type: DataType!, $query: String!) {
  deleteMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	enableImageSearchMutation = `mutation {
  enableImageSearch
}`

	disableImageSearchMutation = `mutation {
  disableImageSearch
}`

	startImageIndexMutation = `mutation startImageIndex($force: Boolean) {
  startImageIndex(force: $force)
}`
)

type Cmd struct {
	LS       LSCmd       `cmd:"" help:"List images."`
	Buckets  BucketsCmd  `cmd:"" help:"List image buckets."`
	Download DownloadCmd `cmd:"" help:"Download an image."`
	Trash    TrashCmd    `cmd:"" help:"Trash images."`
	Restore  RestoreCmd  `cmd:"" help:"Restore images from trash."`
	Delete   DeleteCmd   `cmd:"" help:"Delete images permanently."`
	Search   SearchCmd   `cmd:"" help:"Manage image search."`
}

type LSCmd struct {
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type BucketsCmd struct{}

type DownloadCmd struct {
	ID  string `arg:"" help:"Image ID."`
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

type SearchCmd struct {
	Status  SearchStatusCmd  `cmd:"" help:"Show image search status."`
	Enable  SearchEnableCmd  `cmd:"" help:"Enable image search."`
	Disable SearchDisableCmd `cmd:"" help:"Disable image search."`
	Index   SearchIndexCmd   `cmd:"" help:"Start indexing images."`
}

type (
	SearchStatusCmd  struct{}
	SearchEnableCmd  struct{}
	SearchDisableCmd struct{}
)

type SearchIndexCmd struct {
	Force bool `help:"Force a full reindex."`
}

type imagesListResponse struct {
	Data struct {
		Images []api.Image `json:"images"`
	} `json:"data"`
}

type imageBucketsResponse struct {
	Data struct {
		MediaBuckets []cmdutil.MediaBucket `json:"mediaBuckets"`
	} `json:"data"`
}

type imageSearchStatusResponse struct {
	Data struct {
		ImageSearchStatus api.ImageSearchStatus `json:"imageSearchStatus"`
	} `json:"data"`
}

type imageMutationResponse struct {
	Data struct {
		TrashMediaItems    cmdutil.MediaMutationResult `json:"trashMediaItems"`
		RestoreMediaItems  cmdutil.MediaMutationResult `json:"restoreMediaItems"`
		DeleteMediaItems   cmdutil.MediaMutationResult `json:"deleteMediaItems"`
		EnableImageSearch  bool                        `json:"enableImageSearch"`
		DisableImageSearch bool                        `json:"disableImageSearch"`
		StartImageIndex    bool                        `json:"startImageIndex"`
	} `json:"data"`
}

func (c *LSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	images, err := listImages(
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

	return printer.PrintList(images)
}

func (c *BucketsCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageBucketsResponse
	if err := apiClient.GraphQL(context.Background(), imageBucketsQuery, map[string]any{
		"type": api.DataTypeImage.ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("query image buckets: %w", err)
	}

	return printer.PrintList(resp.Data.MediaBuckets)
}

func (c *DownloadCmd) Run(cli *cmdutil.CLIContext, apiClient *client.Client, printer output.Printer) error {
	reader, err := client.DownloadFile(context.Background(), apiClient, "", c.ID)
	if err != nil {
		return fmt.Errorf("download image: %w", err)
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
			imgDownloadProgressModel{bar: progress.New(progress.WithDefaultGradient()), label: c.Out},
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
					prog.Send(imgDownloadProgressMsg(received))
				}
				if readErr != nil {
					break
				}
			}
			prog.Send(imgDownloadDoneMsg{})
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
	result, err := runImageMediaMutation(context.Background(), apiClient, trashImagesMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *RestoreCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runImageMediaMutation(context.Background(), apiClient, restoreImagesMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *DeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runImageMediaMutation(context.Background(), apiClient, deleteImagesMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *SearchStatusCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageSearchStatusResponse
	if err := apiClient.GraphQL(context.Background(), imageSearchStatusQuery, nil, &resp); err != nil {
		return fmt.Errorf("query image search status: %w", err)
	}

	return printer.Print(resp.Data.ImageSearchStatus)
}

func (c *SearchEnableCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(context.Background(), enableImageSearchMutation, nil, &resp); err != nil {
		return fmt.Errorf("enable image search: %w", err)
	}
	if !resp.Data.EnableImageSearch {
		return errors.New("enable image search: mutation returned false")
	}

	return nil
}

func (c *SearchDisableCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(context.Background(), disableImageSearchMutation, nil, &resp); err != nil {
		return fmt.Errorf("disable image search: %w", err)
	}
	if !resp.Data.DisableImageSearch {
		return errors.New("disable image search: mutation returned false")
	}

	return nil
}

func (c *SearchIndexCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(context.Background(), startImageIndexMutation, map[string]any{
		"force": c.Force,
	}, &resp); err != nil {
		return fmt.Errorf("start image index: %w", err)
	}
	if !resp.Data.StartImageIndex {
		return errors.New("start image index: mutation returned false")
	}

	return nil
}

func listImages(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Image, error) {
	if sortBy == "" {
		sortBy = api.FileSortByDateDesc
	}

	if limit > 0 {
		return fetchImagesPage(ctx, apiClient, query, sortBy, offset, limit)
	}

	images := make([]api.Image, 0, cmdutil.FilesPageSize)
	currentOffset := offset
	for {
		page, err := fetchImagesPage(ctx, apiClient, query, sortBy, currentOffset, cmdutil.FilesPageSize)
		if err != nil {
			return nil, err
		}

		images = append(images, page...)
		if len(page) < cmdutil.FilesPageSize {
			return images, nil
		}

		currentOffset += len(page)
	}
}

func fetchImagesPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Image, error) {
	var resp imagesListResponse
	if err := apiClient.GraphQL(ctx, imagesQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
		"sortBy": sortBy.ToGraphQL(),
	}, &resp); err != nil {
		return nil, fmt.Errorf("query images: %w", err)
	}

	return resp.Data.Images, nil
}

func runImageMediaMutation(
	ctx context.Context,
	apiClient *client.Client,
	mutation string,
	query string,
) (cmdutil.MediaMutationResult, error) {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(ctx, mutation, map[string]any{
		"query": query,
		"type":  api.DataTypeImage.ToGraphQL(),
	}, &resp); err != nil {
		return cmdutil.MediaMutationResult{}, fmt.Errorf("run image media mutation: %w", err)
	}

	switch mutation {
	case trashImagesMutation:
		return resp.Data.TrashMediaItems, nil
	case restoreImagesMutation:
		return resp.Data.RestoreMediaItems, nil
	case deleteImagesMutation:
		return resp.Data.DeleteMediaItems, nil
	default:
		return cmdutil.MediaMutationResult{}, errors.New("run image media mutation: unknown mutation")
	}
}

type (
	imgDownloadProgressMsg int64
	imgDownloadDoneMsg     struct{}
)

type imgDownloadProgressModel struct {
	bar      progress.Model
	label    string
	received int64
}

func (m imgDownloadProgressModel) Init() tea.Cmd { return nil }

func (m imgDownloadProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case imgDownloadProgressMsg:
		m.received = int64(v)
		return m, nil
	case imgDownloadDoneMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m imgDownloadProgressModel) View() string {
	const cycleBytes = 1 << 20
	pct := float64(m.received%cycleBytes) / float64(cycleBytes)
	return m.label + " " + m.bar.ViewAs(pct) + fmt.Sprintf(" %d B", m.received) + "\n"
}
