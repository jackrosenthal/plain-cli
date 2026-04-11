package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
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

type videosListResponse struct {
	Data struct {
		Videos []api.Video `json:"videos"`
	} `json:"data"`
}

type videoBucketsResponse struct {
	Data struct {
		MediaBuckets []mediaBucket `json:"mediaBuckets"`
	} `json:"data"`
}

type videoMutationResponse struct {
	Data struct {
		TrashMediaItems   mediaMutationResult `json:"trashMediaItems"`
		RestoreMediaItems mediaMutationResult `json:"restoreMediaItems"`
		DeleteMediaItems  mediaMutationResult `json:"deleteMediaItems"`
	} `json:"data"`
}

func (c *VideosLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
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

func (c *VideosBucketsCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp videoBucketsResponse
	if err := apiClient.GraphQL(context.Background(), videoBucketsQuery, map[string]any{
		"type": api.DataTypeVideo.ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("query video buckets: %w", err)
	}

	return printer.PrintList(resp.Data.MediaBuckets)
}

func (c *VideosDownloadCmd) Run(cli *CLI, apiClient *client.Client, printer output.Printer) error {
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
		Message: fmt.Sprintf("Downloaded video to %s.", c.Out),
	})
}

func (c *VideosTrashCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runVideoMediaMutation(context.Background(), apiClient, trashVideosMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *VideosRestoreCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runVideoMediaMutation(context.Background(), apiClient, restoreVideosMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *VideosDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
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

	videos := make([]api.Video, 0, filesPageSize)
	currentOffset := offset
	for {
		page, err := fetchVideosPage(ctx, apiClient, query, sortBy, currentOffset, filesPageSize)
		if err != nil {
			return nil, err
		}

		videos = append(videos, page...)
		if len(page) < filesPageSize {
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
) (mediaMutationResult, error) {
	var resp videoMutationResponse
	if err := apiClient.GraphQL(ctx, mutation, map[string]any{
		"query": query,
		"type":  api.DataTypeVideo.ToGraphQL(),
	}, &resp); err != nil {
		return mediaMutationResult{}, fmt.Errorf("run video media mutation: %w", err)
	}

	switch mutation {
	case trashVideosMutation:
		return resp.Data.TrashMediaItems, nil
	case restoreVideosMutation:
		return resp.Data.RestoreMediaItems, nil
	case deleteVideosMutation:
		return resp.Data.DeleteMediaItems, nil
	default:
		return mediaMutationResult{}, errors.New("run video media mutation: unknown mutation")
	}
}
