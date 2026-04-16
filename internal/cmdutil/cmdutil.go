package cmdutil

import (
	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

// CLIContext holds the top-level CLI flags that subcommands may need.
type CLIContext struct {
	Host     string
	Token    string
	ClientID string
	Output   output.Format
}

// MediaBucket represents a media bucket.
type MediaBucket struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	ItemCount int      `json:"itemCount"`
	TopItems  []string `json:"topItems"`
}

// MediaMutationResult is returned from media trash/restore/delete mutations.
type MediaMutationResult struct {
	Type  string `json:"type"`
	Query string `json:"query"`
}

// FilesPageSize is the default page size for paginated file/media queries.
const FilesPageSize = 100

// TagNames extracts the Name field from a slice of api.Tag.
func TagNames(tags []api.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names
}
