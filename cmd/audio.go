package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	audioQuery = `query audios($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  audios(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    id
    title
    artist
    path
    duration
    size
    bucketId
    albumFileId
    createdAt
    updatedAt
    tags {
      id
      name
      count
    }
  }
}`

	playAudioMutation = `mutation playAudio($path: String!) {
  playAudio(path: $path) {
    title
    artist
    path
    duration
  }
}`

	updateAudioPlayModeMutation = `mutation updateAudioPlayMode($mode: MediaPlayMode!) {
  updateAudioPlayMode(mode: $mode)
}`

	trashAudioMutation = `mutation trashMediaItems($type: DataType!, $query: String!) {
  trashMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	restoreAudioMutation = `mutation restoreMediaItems($type: DataType!, $query: String!) {
  restoreMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	deleteAudioMutation = `mutation deleteMediaItems($type: DataType!, $query: String!) {
  deleteMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	audioPlaylistQuery = `query {
  app {
    audios {
      title
      artist
      path
      duration
    }
  }
}`

	addPlaylistAudiosMutation = `mutation addPlaylistAudios($query: String!) {
  addPlaylistAudios(query: $query)
}`

	deletePlaylistAudioMutation = `mutation deletePlaylistAudio($path: String!) {
  deletePlaylistAudio(path: $path)
}`

	clearAudioPlaylistMutation = `mutation {
  clearAudioPlaylist
}`

	reorderPlaylistAudiosMutation = `mutation reorderPlaylistAudios($paths: [String!]!) {
  reorderPlaylistAudios(paths: $paths)
}`
)

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

type audioListResponse struct {
	Data struct {
		Audios []api.Audio `json:"audios"`
	} `json:"data"`
}

type playAudioResponse struct {
	Data struct {
		PlayAudio api.PlaylistAudio `json:"playAudio"`
	} `json:"data"`
}

type audioPlaylistResponse struct {
	Data struct {
		App struct {
			Audios []api.PlaylistAudio `json:"audios"`
		} `json:"app"`
	} `json:"data"`
}

type audioMutationResponse struct {
	Data struct {
		UpdateAudioPlayMode bool                `json:"updateAudioPlayMode"`
		TrashMediaItems     mediaMutationResult `json:"trashMediaItems"`
		RestoreMediaItems   mediaMutationResult `json:"restoreMediaItems"`
		DeleteMediaItems    mediaMutationResult `json:"deleteMediaItems"`
		AddPlaylistAudios   bool                `json:"addPlaylistAudios"`
		DeletePlaylistAudio bool                `json:"deletePlaylistAudio"`
		ClearAudioPlaylist  bool                `json:"clearAudioPlaylist"`
		ReorderPlaylist     bool                `json:"reorderPlaylistAudios"`
	} `json:"data"`
}

func (c *AudioLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	audios, err := listAudios(
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

	return printer.PrintList(audios)
}

func (c *AudioPlayCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp playAudioResponse
	if err := apiClient.GraphQL(context.Background(), playAudioMutation, map[string]any{
		"path": c.Path,
	}, &resp); err != nil {
		return fmt.Errorf("play audio: %w", err)
	}

	return printer.Print(resp.Data.PlayAudio)
}

func (c *AudioModeCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp audioMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateAudioPlayModeMutation, map[string]any{
		"mode": api.MediaPlayMode(c.Mode).ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("update audio play mode: %w", err)
	}
	if !resp.Data.UpdateAudioPlayMode {
		return errors.New("update audio play mode: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Playback mode set to %s.", c.Mode),
	})
}

func (c *AudioTrashCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runAudioMediaMutation(context.Background(), apiClient, trashAudioMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *AudioRestoreCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runAudioMediaMutation(context.Background(), apiClient, restoreAudioMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *AudioDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	result, err := runAudioMediaMutation(context.Background(), apiClient, deleteAudioMutation, c.Query)
	if err != nil {
		return err
	}

	return printer.Print(result)
}

func (c *AudioPlaylistLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp audioPlaylistResponse
	if err := apiClient.GraphQL(context.Background(), audioPlaylistQuery, nil, &resp); err != nil {
		return fmt.Errorf("query audio playlist: %w", err)
	}

	return printer.PrintList(resp.Data.App.Audios)
}

func (c *AudioPlaylistAddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp audioMutationResponse
	if err := apiClient.GraphQL(context.Background(), addPlaylistAudiosMutation, map[string]any{
		"query": c.Query,
	}, &resp); err != nil {
		return fmt.Errorf("add playlist audios: %w", err)
	}
	if !resp.Data.AddPlaylistAudios {
		return errors.New("add playlist audios: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "Audio added to playlist.",
	})
}

func (c *AudioPlaylistRemoveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp audioMutationResponse
	if err := apiClient.GraphQL(context.Background(), deletePlaylistAudioMutation, map[string]any{
		"path": c.Path,
	}, &resp); err != nil {
		return fmt.Errorf("remove playlist audio: %w", err)
	}
	if !resp.Data.DeletePlaylistAudio {
		return errors.New("remove playlist audio: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "Audio removed from playlist.",
	})
}

func (c *AudioPlaylistClearCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp audioMutationResponse
	if err := apiClient.GraphQL(context.Background(), clearAudioPlaylistMutation, nil, &resp); err != nil {
		return fmt.Errorf("clear audio playlist: %w", err)
	}
	if !resp.Data.ClearAudioPlaylist {
		return errors.New("clear audio playlist: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: "Playlist cleared.",
	})
}

func (c *AudioPlaylistReorderCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp audioMutationResponse
	if err := apiClient.GraphQL(context.Background(), reorderPlaylistAudiosMutation, map[string]any{
		"paths": c.Paths,
	}, &resp); err != nil {
		return fmt.Errorf("reorder playlist audios: %w", err)
	}
	if !resp.Data.ReorderPlaylist {
		return errors.New("reorder playlist audios: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Reordered %d playlist item(s).", len(c.Paths)),
	})
}

func listAudios(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Audio, error) {
	if sortBy == "" {
		sortBy = api.FileSortByDateDesc
	}

	if limit > 0 {
		return fetchAudiosPage(ctx, apiClient, query, sortBy, offset, limit)
	}

	audios := make([]api.Audio, 0, filesPageSize)
	currentOffset := offset
	for {
		page, err := fetchAudiosPage(ctx, apiClient, query, sortBy, currentOffset, filesPageSize)
		if err != nil {
			return nil, err
		}

		audios = append(audios, page...)
		if len(page) < filesPageSize {
			return audios, nil
		}

		currentOffset += len(page)
	}
}

func fetchAudiosPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Audio, error) {
	var resp audioListResponse
	if err := apiClient.GraphQL(ctx, audioQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
		"sortBy": sortBy.ToGraphQL(),
	}, &resp); err != nil {
		return nil, fmt.Errorf("query audios: %w", err)
	}

	return resp.Data.Audios, nil
}

func runAudioMediaMutation(
	ctx context.Context,
	apiClient *client.Client,
	mutation string,
	query string,
) (mediaMutationResult, error) {
	var resp audioMutationResponse
	if err := apiClient.GraphQL(ctx, mutation, map[string]any{
		"query": query,
		"type":  api.DataTypeAudio.ToGraphQL(),
	}, &resp); err != nil {
		return mediaMutationResult{}, fmt.Errorf("run audio media mutation: %w", err)
	}

	switch mutation {
	case trashAudioMutation:
		return resp.Data.TrashMediaItems, nil
	case restoreAudioMutation:
		return resp.Data.RestoreMediaItems, nil
	case deleteAudioMutation:
		return resp.Data.DeleteMediaItems, nil
	default:
		return mediaMutationResult{}, errors.New("run audio media mutation: unknown mutation")
	}
}
