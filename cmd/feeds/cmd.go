package feeds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/cmdutil"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	feedsEntriesPageSize = 100

	feedsQuery = `query {
  feeds {
    id
    name
    url
    fetchContent
    createdAt
    updatedAt
  }
}`

	createFeedMutation = `mutation createFeed($url: String!, $fetchContent: Boolean!) {
  createFeed(url: $url, fetchContent: $fetchContent) {
    id
    name
    url
    fetchContent
    createdAt
    updatedAt
  }
}`

	updateFeedMutation = `mutation updateFeed($id: ID!, $name: String!, $fetchContent: Boolean!) {
  updateFeed(id: $id, name: $name, fetchContent: $fetchContent) {
    id
    name
    url
    fetchContent
    createdAt
    updatedAt
  }
}`

	deleteFeedMutation = `mutation deleteFeed($id: ID!) {
  deleteFeed(id: $id)
}`

	syncFeedsMutation = `mutation syncFeeds($id: ID) {
  syncFeeds(id: $id)
}`

	importFeedsMutation = `mutation importFeeds($content: String!) {
  importFeeds(content: $content)
}`

	exportFeedsMutation = `mutation {
  exportFeeds
}`

	feedEntriesQuery = `query feedEntries($offset: Int!, $limit: Int!, $query: String!) {
  items: feedEntries(offset: $offset, limit: $limit, query: $query) {
    id
    title
    url
    image
    author
    feedId
    rawId
    publishedAt
    createdAt
    updatedAt
    tags {
      id
      name
    }
  }
  total: feedEntryCount(query: $query)
}`

	feedEntryQuery = `query feedEntry($id: ID!) {
  feedEntry(id: $id) {
    id
    title
    url
    image
    author
    description
    content
    feedId
    rawId
    publishedAt
    createdAt
    updatedAt
    tags {
      id
      name
      count
    }
    feed {
      id
      name
      url
      fetchContent
      createdAt
      updatedAt
    }
  }
}`

	deleteFeedEntriesMutation = `mutation deleteFeedEntries($query: String!) {
  deleteFeedEntries(query: $query)
}`

	saveFeedEntriesToNotesMutation = `mutation saveFeedEntriesToNotes($query: String!) {
  saveFeedEntriesToNotes(query: $query)
}`
)

type Cmd struct {
	LS      LSCmd      `cmd:"" help:"List feeds."`
	Add     AddCmd     `cmd:"" help:"Add a feed."`
	Update  UpdateCmd  `cmd:"" help:"Update a feed."`
	Delete  DeleteCmd  `cmd:"" help:"Delete a feed."`
	Sync    SyncCmd    `cmd:"" help:"Sync feeds."`
	Import  ImportCmd  `cmd:"" help:"Import feeds from OPML."`
	Export  ExportCmd  `cmd:"" help:"Export feeds as OPML."`
	Entries EntriesCmd `cmd:"" help:"Manage feed entries."`
}

type LSCmd struct{}

type AddCmd struct {
	URL          string `arg:"" help:"Feed URL."`
	FetchContent bool   `help:"Fetch article content."`
}

type UpdateCmd struct {
	ID           string `arg:"" help:"Feed ID."`
	Name         string `help:"Feed name." required:""`
	FetchContent bool   `help:"Fetch article content."`
}

type DeleteCmd struct {
	ID string `arg:"" help:"Feed ID."`
}

type SyncCmd struct {
	ID string `help:"Specific feed ID to sync."`
}

type ImportCmd struct {
	OpmlFile string `arg:"" help:"Local OPML file path."`
}

type ExportCmd struct{}

type EntriesCmd struct {
	LS          EntriesLSCmd          `cmd:"" help:"List feed entries."`
	Get         EntriesGetCmd         `cmd:"" help:"Get a feed entry by ID."`
	Delete      EntriesDeleteCmd      `cmd:"" help:"Delete feed entries."`
	SaveToNotes EntriesSaveToNotesCmd `cmd:"" help:"Save feed entries to notes."`
}

type EntriesLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type EntriesGetCmd struct {
	ID string `arg:"" help:"Feed entry ID."`
}

type EntriesDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type EntriesSaveToNotesCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type feedsResponse struct {
	Data struct {
		Feeds []api.Feed `json:"feeds"`
	} `json:"data"`
}

type feedMutationResponse struct {
	Data struct {
		CreateFeed             api.Feed `json:"createFeed"`
		UpdateFeed             api.Feed `json:"updateFeed"`
		DeleteFeed             bool     `json:"deleteFeed"`
		SyncFeeds              bool     `json:"syncFeeds"`
		ImportFeeds            bool     `json:"importFeeds"`
		ExportFeeds            string   `json:"exportFeeds"`
		DeleteFeedEntries      bool     `json:"deleteFeedEntries"`
		SaveFeedEntriesToNotes bool     `json:"saveFeedEntriesToNotes"`
	} `json:"data"`
}

type feedEntriesResponse struct {
	Data struct {
		Items []feedEntryListRecord `json:"items"`
		Total int                   `json:"total"`
	} `json:"data"`
}

type feedEntryResponse struct {
	Data struct {
		FeedEntry feedEntryDetail `json:"feedEntry"`
	} `json:"data"`
}

type feedEntryListRecord struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Image       string    `json:"image"`
	Author      string    `json:"author"`
	FeedID      string    `json:"feedId"`
	RawID       string    `json:"rawId"`
	PublishedAt string    `json:"publishedAt"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	Tags        []api.Tag `json:"tags"`
}

type feedEntryListItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Image       string   `json:"image"`
	Author      string   `json:"author"`
	FeedID      string   `json:"feedId"`
	RawID       string   `json:"rawId"`
	PublishedAt string   `json:"publishedAt"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Tags        []string `json:"tags"`
}

type feedEntryDetail struct {
	api.FeedEntry
	Feed api.Feed `json:"feed"`
}

func (c *LSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp feedsResponse
	if err := apiClient.GraphQL(context.Background(), feedsQuery, nil, &resp); err != nil {
		return fmt.Errorf("query feeds: %w", err)
	}

	return printer.PrintList(resp.Data.Feeds)
}

func (c *AddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp feedMutationResponse
	if err := apiClient.GraphQL(context.Background(), createFeedMutation, map[string]any{
		"fetchContent": c.FetchContent,
		"url":          c.URL,
	}, &resp); err != nil {
		return fmt.Errorf("create feed: %w", err)
	}

	return printer.Print(resp.Data.CreateFeed)
}

func (c *UpdateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp feedMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateFeedMutation, map[string]any{
		"fetchContent": c.FetchContent,
		"id":           c.ID,
		"name":         c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("update feed: %w", err)
	}

	return printer.Print(resp.Data.UpdateFeed)
}

func (c *DeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp feedMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteFeedMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("delete feed: %w", err)
	}
	if !resp.Data.DeleteFeed {
		return errors.New("delete feed: mutation returned false")
	}

	return nil
}

func (c *SyncCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var id any
	if c.ID != "" {
		id = c.ID
	}

	var resp feedMutationResponse
	if err := apiClient.GraphQL(context.Background(), syncFeedsMutation, map[string]any{
		"id": id,
	}, &resp); err != nil {
		return fmt.Errorf("sync feeds: %w", err)
	}
	if !resp.Data.SyncFeeds {
		return errors.New("sync feeds: mutation returned false")
	}

	return nil
}

func (c *ImportCmd) Run(apiClient *client.Client, printer output.Printer) error {
	content, err := os.ReadFile(c.OpmlFile)
	if err != nil {
		return fmt.Errorf("read OPML file: %w", err)
	}

	var resp feedMutationResponse
	if err := apiClient.GraphQL(context.Background(), importFeedsMutation, map[string]any{
		"content": string(content),
	}, &resp); err != nil {
		return fmt.Errorf("import feeds: %w", err)
	}
	if !resp.Data.ImportFeeds {
		return errors.New("import feeds: mutation returned false")
	}

	return nil
}

func (c *ExportCmd) Run(apiClient *client.Client) error {
	var resp feedMutationResponse
	if err := apiClient.GraphQL(context.Background(), exportFeedsMutation, nil, &resp); err != nil {
		return fmt.Errorf("export feeds: %w", err)
	}

	_, err := io.WriteString(os.Stdout, resp.Data.ExportFeeds)
	return err
}

func (c *EntriesLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	items, err := listFeedEntries(context.Background(), apiClient, c.Query, c.Offset, c.Limit)
	if err != nil {
		return err
	}

	return printer.PrintList(items)
}

func (c *EntriesGetCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp feedEntryResponse
	if err := apiClient.GraphQL(context.Background(), feedEntryQuery, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("query feed entry: %w", err)
	}

	return printer.Print(resp.Data.FeedEntry)
}

func (c *EntriesDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	if err := runFeedsBoolMutation(
		context.Background(),
		apiClient,
		deleteFeedEntriesMutation,
		"deleteFeedEntries",
		c.Query,
	); err != nil {
		return err
	}

	return nil
}

func (c *EntriesSaveToNotesCmd) Run(apiClient *client.Client, printer output.Printer) error {
	if err := runFeedsBoolMutation(
		context.Background(),
		apiClient,
		saveFeedEntriesToNotesMutation,
		"saveFeedEntriesToNotes",
		c.Query,
	); err != nil {
		return err
	}

	return nil
}

func listFeedEntries(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]feedEntryListItem, error) {
	if limit > 0 {
		return fetchFeedEntriesPage(ctx, apiClient, query, offset, limit)
	}

	items := make([]feedEntryListItem, 0, feedsEntriesPageSize)
	currentOffset := offset
	for {
		page, err := fetchFeedEntriesPage(ctx, apiClient, query, currentOffset, feedsEntriesPageSize)
		if err != nil {
			return nil, err
		}

		items = append(items, page...)
		if len(page) < feedsEntriesPageSize {
			return items, nil
		}

		currentOffset += len(page)
	}
}

func fetchFeedEntriesPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]feedEntryListItem, error) {
	var resp feedEntriesResponse
	if err := apiClient.GraphQL(ctx, feedEntriesQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
	}, &resp); err != nil {
		return nil, fmt.Errorf("query feed entries: %w", err)
	}

	return displayFeedEntries(resp.Data.Items), nil
}

func runFeedsBoolMutation(
	ctx context.Context,
	apiClient *client.Client,
	mutation string,
	field string,
	query string,
) error {
	var resp feedMutationResponse
	if err := apiClient.GraphQL(ctx, mutation, map[string]any{
		"query": query,
	}, &resp); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}

	var ok bool
	switch field {
	case "deleteFeedEntries":
		ok = resp.Data.DeleteFeedEntries
	case "saveFeedEntriesToNotes":
		ok = resp.Data.SaveFeedEntriesToNotes
	default:
		return fmt.Errorf("%s: unsupported mutation field", field)
	}
	if !ok {
		return fmt.Errorf("%s: %w", field, errors.New("mutation returned false"))
	}

	return nil
}

func displayFeedEntries(records []feedEntryListRecord) []feedEntryListItem {
	items := make([]feedEntryListItem, 0, len(records))
	for _, record := range records {
		items = append(items, feedEntryListItem{
			ID:          record.ID,
			Title:       record.Title,
			URL:         record.URL,
			Image:       record.Image,
			Author:      record.Author,
			FeedID:      record.FeedID,
			RawID:       record.RawID,
			PublishedAt: record.PublishedAt,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
			Tags:        cmdutil.TagNames(record.Tags),
		})
	}

	return items
}
