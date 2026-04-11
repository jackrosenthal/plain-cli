package cmd

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
	Pinned    bool   `help:"Pin the bookmark."`
	SortOrder int    `name:"sort-order" help:"Sort order."`
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
	Collapsed bool   `help:"Collapse the group in the UI."`
	SortOrder int    `name:"sort-order" help:"Sort order."`
}

type BookmarksGroupsDeleteCmd struct {
	ID string `arg:"" help:"Group ID."`
}
