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

type imagesListResponse struct {
	Data struct {
		Images []api.Image `json:"images"`
	} `json:"data"`
}

type imageBucketsResponse struct {
	Data struct {
		MediaBuckets []mediaBucket `json:"mediaBuckets"`
	} `json:"data"`
}

type imageSearchStatusResponse struct {
	Data struct {
		ImageSearchStatus api.ImageSearchStatus `json:"imageSearchStatus"`
	} `json:"data"`
}

type imageMutationResponse struct {
	Data struct {
		TrashMediaItems    mediaMutationResult `json:"trashMediaItems"`
		RestoreMediaItems  mediaMutationResult `json:"restoreMediaItems"`
		DeleteMediaItems   mediaMutationResult `json:"deleteMediaItems"`
		EnableImageSearch  bool                `json:"enableImageSearch"`
		DisableImageSearch bool                `json:"disableImageSearch"`
		StartImageIndex    bool                `json:"startImageIndex"`
	} `json:"data"`
}

type mediaBucket struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	ItemCount int      `json:"itemCount"`
	TopItems  []string `json:"topItems"`
}

type mediaMutationResult struct {
	Type  string `json:"type"`
	Query string `json:"query"`
}

func (c *ImagesLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
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

func (c *ImagesBucketsCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageBucketsResponse
	if err := apiClient.GraphQL(context.Background(), imageBucketsQuery, map[string]any{
		"type": api.DataTypeImage.ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("query image buckets: %w", err)
	}

	return printer.PrintList(resp.Data.MediaBuckets)
}

func (c *ImagesDownloadCmd) Run(cli *CLI, apiClient *client.Client, printer output.Printer) error {
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
		Message: fmt.Sprintf("Downloaded image to %s.", c.Out),
	})
}

func (c *ImagesTrashCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runImageMediaMutation(context.Background(), apiClient, trashImagesMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *ImagesRestoreCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runImageMediaMutation(context.Background(), apiClient, restoreImagesMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *ImagesDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runImageMediaMutation(context.Background(), apiClient, deleteImagesMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *ImagesSearchStatusCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageSearchStatusResponse
	if err := apiClient.GraphQL(context.Background(), imageSearchStatusQuery, nil, &resp); err != nil {
		return fmt.Errorf("query image search status: %w", err)
	}

	return printer.Print(resp.Data.ImageSearchStatus)
}

func (c *ImagesSearchEnableCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(context.Background(), enableImageSearchMutation, nil, &resp); err != nil {
		return fmt.Errorf("enable image search: %w", err)
	}
	if !resp.Data.EnableImageSearch {
		return errors.New("enable image search: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "Image search enabled.",
	})
}

func (c *ImagesSearchDisableCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(context.Background(), disableImageSearchMutation, nil, &resp); err != nil {
		return fmt.Errorf("disable image search: %w", err)
	}
	if !resp.Data.DisableImageSearch {
		return errors.New("disable image search: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "Image search disabled.",
	})
}

func (c *ImagesSearchIndexCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(context.Background(), startImageIndexMutation, map[string]any{
		"force": c.Force,
	}, &resp); err != nil {
		return fmt.Errorf("start image index: %w", err)
	}
	if !resp.Data.StartImageIndex {
		return errors.New("start image index: mutation returned false")
	}

	message := "Image indexing started."
	if c.Force {
		message = "Forced image indexing started."
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: message,
	})
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

	images := make([]api.Image, 0, filesPageSize)
	currentOffset := offset
	for {
		page, err := fetchImagesPage(ctx, apiClient, query, sortBy, currentOffset, filesPageSize)
		if err != nil {
			return nil, err
		}

		images = append(images, page...)
		if len(page) < filesPageSize {
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
) (mediaMutationResult, error) {
	var resp imageMutationResponse
	if err := apiClient.GraphQL(ctx, mutation, map[string]any{
		"query": query,
		"type":  api.DataTypeImage.ToGraphQL(),
	}, &resp); err != nil {
		return mediaMutationResult{}, fmt.Errorf("run image media mutation: %w", err)
	}

	switch mutation {
	case trashImagesMutation:
		return resp.Data.TrashMediaItems, nil
	case restoreImagesMutation:
		return resp.Data.RestoreMediaItems, nil
	case deleteImagesMutation:
		return resp.Data.DeleteMediaItems, nil
	default:
		return mediaMutationResult{}, errors.New("run image media mutation: unknown mutation")
	}
}
