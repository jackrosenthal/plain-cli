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
	bookmarksQuery = `query {
  bookmarks {
    id
    url
    title
    faviconPath
    groupId
    pinned
    clickCount
    lastClickedAt
    sortOrder
    createdAt
    updatedAt
  }
}`

	addBookmarksMutation = `mutation addBookmarks($urls: [String!]!, $groupId: String!) {
  addBookmarks(urls: $urls, groupId: $groupId) {
    id
    url
    title
    faviconPath
    groupId
    pinned
    clickCount
    lastClickedAt
    sortOrder
    createdAt
    updatedAt
  }
}`

	updateBookmarkMutation = `mutation updateBookmark($id: ID!, $input: BookmarkInput!) {
  updateBookmark(id: $id, input: $input) {
    id
    url
    title
    faviconPath
    groupId
    pinned
    clickCount
    lastClickedAt
    sortOrder
    createdAt
    updatedAt
  }
}`

	deleteBookmarksMutation = `mutation deleteBookmarks($ids: [ID!]!) {
  deleteBookmarks(ids: $ids)
}`

	bookmarkGroupsQuery = `query {
  bookmarkGroups {
    id
    name
    collapsed
    sortOrder
    createdAt
    updatedAt
  }
}`

	createBookmarkGroupMutation = `mutation createBookmarkGroup($name: String!) {
  createBookmarkGroup(name: $name) {
    id
    name
    collapsed
    sortOrder
    createdAt
    updatedAt
  }
}`

	updateBookmarkGroupMutation = `mutation updateBookmarkGroup($id: ID!, $name: String!, $collapsed: Boolean!, $sortOrder: Int!) {
  updateBookmarkGroup(id: $id, name: $name, collapsed: $collapsed, sortOrder: $sortOrder) {
    id
    name
    collapsed
    sortOrder
    createdAt
    updatedAt
  }
}`

	deleteBookmarkGroupMutation = `mutation deleteBookmarkGroup($id: ID!) {
  deleteBookmarkGroup(id: $id)
}`
)

type BookmarksCmd struct {
	LS     BookmarksLSCmd     `cmd:"" help:"List bookmarks."`
	Add    BookmarksAddCmd    `cmd:"" help:"Add bookmarks."`
	Update BookmarksUpdateCmd `cmd:"" help:"Update a bookmark."`
	Delete BookmarksDeleteCmd `cmd:"" help:"Delete bookmarks."`
	Groups BookmarksGroupsCmd `cmd:"" help:"Manage bookmark groups."`
}

type BookmarksLSCmd struct{}

type BookmarksAddCmd struct {
	URLs    []string `arg:"" name:"urls" help:"Bookmark URLs to add."`
	GroupID string   `name:"group-id" help:"Destination group ID." required:""`
}

type BookmarksUpdateCmd struct {
	ID        string `arg:"" help:"Bookmark ID."`
	URL       string `help:"Bookmark URL."`
	Title     string `help:"Bookmark title."`
	GroupID   string `name:"group-id" help:"Bookmark group ID."`
	Pinned    *bool  `help:"Pin or unpin the bookmark."`
	SortOrder *int   `name:"sort-order" help:"Sort order."`
}

type BookmarksDeleteCmd struct {
	IDs []string `arg:"" name:"ids" help:"Bookmark IDs to delete."`
}

type BookmarksGroupsCmd struct {
	LS     BookmarksGroupsLSCmd     `cmd:"" help:"List bookmark groups."`
	Create BookmarksGroupsCreateCmd `cmd:"" help:"Create a bookmark group."`
	Update BookmarksGroupsUpdateCmd `cmd:"" help:"Update a bookmark group."`
	Delete BookmarksGroupsDeleteCmd `cmd:"" help:"Delete a bookmark group."`
}

type BookmarksGroupsLSCmd struct{}

type BookmarksGroupsCreateCmd struct {
	Name string `arg:"" help:"Group name."`
}

type BookmarksGroupsUpdateCmd struct {
	ID        string `arg:"" help:"Group ID."`
	Name      string `help:"Group name."`
	Collapsed *bool  `help:"Collapse or expand the group in the UI."`
	SortOrder *int   `name:"sort-order" help:"Sort order."`
}

type BookmarksGroupsDeleteCmd struct {
	ID string `arg:"" help:"Group ID."`
}

type bookmarksResponse struct {
	Data struct {
		Bookmarks []api.Bookmark `json:"bookmarks"`
	} `json:"data"`
}

type bookmarkMutationResponse struct {
	Data struct {
		AddBookmarks    []api.Bookmark `json:"addBookmarks"`
		UpdateBookmark  api.Bookmark   `json:"updateBookmark"`
		DeleteBookmarks bool           `json:"deleteBookmarks"`
	} `json:"data"`
}

type bookmarkGroupsResponse struct {
	Data struct {
		BookmarkGroups []api.BookmarkGroup `json:"bookmarkGroups"`
	} `json:"data"`
}

type bookmarkGroupMutationResponse struct {
	Data struct {
		CreateBookmarkGroup api.BookmarkGroup `json:"createBookmarkGroup"`
		UpdateBookmarkGroup api.BookmarkGroup `json:"updateBookmarkGroup"`
		DeleteBookmarkGroup bool              `json:"deleteBookmarkGroup"`
	} `json:"data"`
}

func (c *BookmarksLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	bookmarks, err := listBookmarks(context.Background(), apiClient)
	if err != nil {
		return err
	}

	return printer.PrintList(bookmarks)
}

func (c *BookmarksAddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp bookmarkMutationResponse
	if err := apiClient.GraphQL(context.Background(), addBookmarksMutation, map[string]any{
		"groupId": c.GroupID,
		"urls":    c.URLs,
	}, &resp); err != nil {
		return fmt.Errorf("add bookmarks: %w", err)
	}

	return printer.PrintList(resp.Data.AddBookmarks)
}

func (c *BookmarksUpdateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	bookmark, err := getBookmark(context.Background(), apiClient, c.ID)
	if err != nil {
		return err
	}

	if c.URL != "" {
		bookmark.URL = c.URL
	}
	if c.Title != "" {
		bookmark.Title = c.Title
	}
	if c.GroupID != "" {
		bookmark.GroupID = c.GroupID
	}
	if c.Pinned != nil {
		bookmark.Pinned = *c.Pinned
	}
	if c.SortOrder != nil {
		bookmark.SortOrder = *c.SortOrder
	}

	var resp bookmarkMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateBookmarkMutation, map[string]any{
		"id": c.ID,
		"input": map[string]any{
			"groupId":   bookmark.GroupID,
			"pinned":    bookmark.Pinned,
			"sortOrder": bookmark.SortOrder,
			"title":     bookmark.Title,
			"url":       bookmark.URL,
		},
	}, &resp); err != nil {
		return fmt.Errorf("update bookmark: %w", err)
	}

	return printer.Print(resp.Data.UpdateBookmark)
}

func (c *BookmarksDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp bookmarkMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteBookmarksMutation, map[string]any{
		"ids": c.IDs,
	}, &resp); err != nil {
		return fmt.Errorf("delete bookmarks: %w", err)
	}
	if !resp.Data.DeleteBookmarks {
		return errors.New("delete bookmarks: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Deleted %d bookmark(s).", len(c.IDs)),
	})
}

func (c *BookmarksGroupsLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	groups, err := listBookmarkGroups(context.Background(), apiClient)
	if err != nil {
		return err
	}

	return printer.PrintList(groups)
}

func (c *BookmarksGroupsCreateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp bookmarkGroupMutationResponse
	if err := apiClient.GraphQL(context.Background(), createBookmarkGroupMutation, map[string]any{
		"name": c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("create bookmark group: %w", err)
	}

	return printer.Print(resp.Data.CreateBookmarkGroup)
}

func (c *BookmarksGroupsUpdateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	group, err := getBookmarkGroup(context.Background(), apiClient, c.ID)
	if err != nil {
		return err
	}

	if c.Name != "" {
		group.Name = c.Name
	}
	if c.Collapsed != nil {
		group.Collapsed = *c.Collapsed
	}
	if c.SortOrder != nil {
		group.SortOrder = *c.SortOrder
	}

	var resp bookmarkGroupMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateBookmarkGroupMutation, map[string]any{
		"collapsed": group.Collapsed,
		"id":        c.ID,
		"name":      group.Name,
		"sortOrder": group.SortOrder,
	}, &resp); err != nil {
		return fmt.Errorf("update bookmark group: %w", err)
	}

	return printer.Print(resp.Data.UpdateBookmarkGroup)
}

func (c *BookmarksGroupsDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp bookmarkGroupMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteBookmarkGroupMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("delete bookmark group: %w", err)
	}
	if !resp.Data.DeleteBookmarkGroup {
		return errors.New("delete bookmark group: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Deleted bookmark group %s.", c.ID),
	})
}

func listBookmarks(ctx context.Context, apiClient *client.Client) ([]api.Bookmark, error) {
	var resp bookmarksResponse
	if err := apiClient.GraphQL(ctx, bookmarksQuery, nil, &resp); err != nil {
		return nil, fmt.Errorf("query bookmarks: %w", err)
	}

	return resp.Data.Bookmarks, nil
}

func getBookmark(ctx context.Context, apiClient *client.Client, id string) (api.Bookmark, error) {
	bookmarks, err := listBookmarks(ctx, apiClient)
	if err != nil {
		return api.Bookmark{}, err
	}

	for _, bookmark := range bookmarks {
		if bookmark.ID == id {
			return bookmark, nil
		}
	}

	return api.Bookmark{}, fmt.Errorf("bookmark %s not found", id)
}

func listBookmarkGroups(ctx context.Context, apiClient *client.Client) ([]api.BookmarkGroup, error) {
	var resp bookmarkGroupsResponse
	if err := apiClient.GraphQL(ctx, bookmarkGroupsQuery, nil, &resp); err != nil {
		return nil, fmt.Errorf("query bookmark groups: %w", err)
	}

	return resp.Data.BookmarkGroups, nil
}

func getBookmarkGroup(ctx context.Context, apiClient *client.Client, id string) (api.BookmarkGroup, error) {
	groups, err := listBookmarkGroups(ctx, apiClient)
	if err != nil {
		return api.BookmarkGroup{}, err
	}

	for _, group := range groups {
		if group.ID == id {
			return group, nil
		}
	}

	return api.BookmarkGroup{}, fmt.Errorf("bookmark group %s not found", id)
}
