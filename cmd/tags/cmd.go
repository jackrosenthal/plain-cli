package tags

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	tagsQuery = `query tags($type: DataType!) {
  tags(type: $type) {
    id
    name
    count
  }
}`

	createTagMutation = `mutation createTag($type: DataType!, $name: String!) {
  createTag(type: $type, name: $name) {
    id
    name
    count
  }
}`

	updateTagMutation = `mutation updateTag($id: ID!, $name: String!) {
  updateTag(id: $id, name: $name) {
    id
    name
    count
  }
}`

	deleteTagMutation = `mutation deleteTag($id: ID!) {
  deleteTag(id: $id)
}`

	addToTagsMutation = `mutation addToTags($type: DataType!, $tagIds: [ID!]!, $query: String!) {
  addToTags(type: $type, tagIds: $tagIds, query: $query)
}`

	removeFromTagsMutation = `mutation removeFromTags($type: DataType!, $tagIds: [ID!]!, $query: String!) {
  removeFromTags(type: $type, tagIds: $tagIds, query: $query)
}`
)

type Cmd struct {
	LS     LSCmd     `cmd:"" help:"List tags for a content type."`
	Create CreateCmd `cmd:"" help:"Create a tag."`
	Update UpdateCmd `cmd:"" help:"Update a tag."`
	Delete DeleteCmd `cmd:"" help:"Delete a tag."`
	Add    AddCmd    `cmd:"" help:"Add items to tags."`
	Remove RemoveCmd `cmd:"" help:"Remove items from tags."`
}

type LSCmd struct {
	Type string `help:"Content type." enum:"image,video,audio,note,feed-entry,call,contact,message,bookmark" required:""`
}

type CreateCmd struct {
	Type string `help:"Content type." enum:"image,video,audio,note,feed-entry,call,contact,message,bookmark" required:""`
	Name string `help:"Tag name." required:""`
}

type UpdateCmd struct {
	ID   string `arg:"" help:"Tag ID."`
	Name string `help:"New tag name." required:""`
}

type DeleteCmd struct {
	ID string `arg:"" help:"Tag ID."`
}

type AddCmd struct {
	Type  string `help:"Content type." enum:"image,video,audio,note,feed-entry,call,contact,message,bookmark" required:""`
	Query string `help:"Selection query." required:""`
	Tags  string `help:"Comma-separated tag IDs." required:""`
}

type RemoveCmd struct {
	Type  string `help:"Content type." enum:"image,video,audio,note,feed-entry,call,contact,message,bookmark" required:""`
	Query string `help:"Selection query." required:""`
	Tags  string `help:"Comma-separated tag IDs." required:""`
}

type tagsResponse struct {
	Data struct {
		Tags []api.Tag `json:"tags"`
	} `json:"data"`
}

type tagMutationResponse struct {
	Data struct {
		CreateTag      api.Tag `json:"createTag"`
		UpdateTag      api.Tag `json:"updateTag"`
		DeleteTag      bool    `json:"deleteTag"`
		AddToTags      bool    `json:"addToTags"`
		RemoveFromTags bool    `json:"removeFromTags"`
	} `json:"data"`
}

func (c *LSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp tagsResponse
	if err := apiClient.GraphQL(context.Background(), tagsQuery, map[string]any{
		"type": api.DataType(c.Type).ToGraphQL(),
	}, &resp); err != nil {
		return fmt.Errorf("query tags: %w", err)
	}

	return printer.PrintList(resp.Data.Tags)
}

func (c *CreateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp tagMutationResponse
	if err := apiClient.GraphQL(context.Background(), createTagMutation, map[string]any{
		"type": api.DataType(c.Type).ToGraphQL(),
		"name": c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("create tag: %w", err)
	}

	return printer.Print(resp.Data.CreateTag)
}

func (c *UpdateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp tagMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateTagMutation, map[string]any{
		"id":   c.ID,
		"name": c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("update tag: %w", err)
	}

	return printer.Print(resp.Data.UpdateTag)
}

func (c *DeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp tagMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteTagMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	if !resp.Data.DeleteTag {
		return errors.New("delete tag: mutation returned false")
	}

	return nil
}

func (c *AddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	tagIDs, err := parseTagIDs(c.Tags)
	if err != nil {
		return err
	}

	var resp tagMutationResponse
	if err := apiClient.GraphQL(context.Background(), addToTagsMutation, map[string]any{
		"type":   api.DataType(c.Type).ToGraphQL(),
		"tagIds": tagIDs,
		"query":  c.Query,
	}, &resp); err != nil {
		return fmt.Errorf("add to tags: %w", err)
	}
	if !resp.Data.AddToTags {
		return errors.New("add to tags: mutation returned false")
	}

	return nil
}

func (c *RemoveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	tagIDs, err := parseTagIDs(c.Tags)
	if err != nil {
		return err
	}

	var resp tagMutationResponse
	if err := apiClient.GraphQL(context.Background(), removeFromTagsMutation, map[string]any{
		"type":   api.DataType(c.Type).ToGraphQL(),
		"tagIds": tagIDs,
		"query":  c.Query,
	}, &resp); err != nil {
		return fmt.Errorf("remove from tags: %w", err)
	}
	if !resp.Data.RemoveFromTags {
		return errors.New("remove from tags: mutation returned false")
	}

	return nil
}

func parseTagIDs(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	tagIDs := make([]string, 0, len(parts))
	for _, part := range parts {
		tagID := strings.TrimSpace(part)
		if tagID == "" {
			continue
		}

		tagIDs = append(tagIDs, tagID)
	}

	if len(tagIDs) == 0 {
		return nil, errors.New("tags must contain at least one tag ID")
	}

	return tagIDs, nil
}
