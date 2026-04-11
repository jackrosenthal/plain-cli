package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	notesPageSize = 100

	notesQuery = `query notes($offset: Int!, $limit: Int!, $query: String!) {
  notes(offset: $offset, limit: $limit, query: $query) {
    id
    title
    deletedAt
    createdAt
    updatedAt
    tags {
      id
      name
    }
  }
}`

	noteQuery = `query note($id: ID!) {
  note(id: $id) {
    id
    title
    content
    deletedAt
    createdAt
    updatedAt
    tags {
      id
      name
      count
    }
  }
}`

	saveNoteMutation = `mutation saveNote($id: ID!, $input: NoteInput!) {
  saveNote(id: $id, input: $input) {
    id
    title
    content
    deletedAt
    createdAt
    updatedAt
    tags {
      id
      name
      count
    }
  }
}`

	trashNotesMutation = `mutation trashNotes($query: String!) {
  trashNotes(query: $query)
}`

	restoreNotesMutation = `mutation restoreNotes($query: String!) {
  restoreNotes(query: $query)
}`

	deleteNotesMutation = `mutation deleteNotes($query: String!) {
  deleteNotes(query: $query)
}`

	exportNotesMutation = `mutation exportNotes($query: String!) {
  exportNotes(query: $query)
}`
)

type NotesCmd struct {
	LS      NotesLSCmd      `cmd:"" help:"List notes."`
	Get     NotesGetCmd     `cmd:"" help:"Get a note by ID."`
	Save    NotesSaveCmd    `cmd:"" help:"Create or update a note."`
	Trash   NotesTrashCmd   `cmd:"" help:"Trash notes."`
	Restore NotesRestoreCmd `cmd:"" help:"Restore notes from trash."`
	Delete  NotesDeleteCmd  `cmd:"" help:"Delete notes permanently."`
	Export  NotesExportCmd  `cmd:"" help:"Export notes."`
}

type NotesLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type NotesGetCmd struct {
	ID string `arg:"" help:"Note ID."`
}

type NotesSaveCmd struct {
	ID      string `help:"Existing note ID."`
	Title   string `help:"Note title." required:""`
	Content string `help:"Note content. Reads stdin when omitted."`
}

type NotesTrashCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type NotesRestoreCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type NotesDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type NotesExportCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type notesListResponse struct {
	Data struct {
		Notes []noteListRecord `json:"notes"`
	} `json:"data"`
}

type noteResponse struct {
	Data struct {
		Note api.Note `json:"note"`
	} `json:"data"`
}

type noteMutationResponse struct {
	Data struct {
		SaveNote     api.Note `json:"saveNote"`
		TrashNotes   bool     `json:"trashNotes"`
		RestoreNotes bool     `json:"restoreNotes"`
		DeleteNotes  bool     `json:"deleteNotes"`
		ExportNotes  string   `json:"exportNotes"`
	} `json:"data"`
}

type noteListItem struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	DeletedAt string   `json:"deletedAt"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	Tags      []string `json:"tags"`
}

type noteListRecord struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	DeletedAt string    `json:"deletedAt"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	Tags      []api.Tag `json:"tags"`
}

type exportedNotes struct {
	Content string `json:"content"`
}

func (c *NotesLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	notes, err := listNotes(context.Background(), apiClient, c.Query, c.Offset, c.Limit)
	if err != nil {
		return err
	}

	return printer.PrintList(notes)
}

func (c *NotesGetCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp noteResponse
	if err := apiClient.GraphQL(context.Background(), noteQuery, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("query note: %w", err)
	}

	return printer.Print(resp.Data.Note)
}

func (c *NotesSaveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	content, err := resolveWriteContent(c.Content)
	if err != nil {
		return err
	}

	id := c.ID
	if id == "" {
		id, err = randomUUID()
		if err != nil {
			return fmt.Errorf("generate note id: %w", err)
		}
	}

	var resp noteMutationResponse
	if err := apiClient.GraphQL(context.Background(), saveNoteMutation, map[string]any{
		"id": id,
		"input": map[string]any{
			"title":   c.Title,
			"content": content,
		},
	}, &resp); err != nil {
		return fmt.Errorf("save note: %w", err)
	}

	return printer.Print(resp.Data.SaveNote)
}

func (c *NotesTrashCmd) Run(apiClient *client.Client, printer output.Printer) error {
	if err := runNotesBoolMutation(context.Background(), apiClient, trashNotesMutation, "trashNotes", c.Query); err != nil {
		return err
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Trashed notes matching %q.", c.Query),
	})
}

func (c *NotesRestoreCmd) Run(apiClient *client.Client, printer output.Printer) error {
	if err := runNotesBoolMutation(context.Background(), apiClient, restoreNotesMutation, "restoreNotes", c.Query); err != nil {
		return err
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Restored notes matching %q.", c.Query),
	})
}

func (c *NotesDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	if err := runNotesBoolMutation(context.Background(), apiClient, deleteNotesMutation, "deleteNotes", c.Query); err != nil {
		return err
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Deleted notes matching %q.", c.Query),
	})
}

func (c *NotesExportCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp noteMutationResponse
	if err := apiClient.GraphQL(context.Background(), exportNotesMutation, map[string]any{
		"query": c.Query,
	}, &resp); err != nil {
		return fmt.Errorf("export notes: %w", err)
	}

	return printer.Print(exportedNotes{Content: resp.Data.ExportNotes})
}

func listNotes(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]noteListItem, error) {
	if limit > 0 {
		return fetchNotesPage(ctx, apiClient, query, offset, limit)
	}

	notes := make([]noteListItem, 0, notesPageSize)
	currentOffset := offset
	for {
		page, err := fetchNotesPage(ctx, apiClient, query, currentOffset, notesPageSize)
		if err != nil {
			return nil, err
		}

		notes = append(notes, page...)
		if len(page) < notesPageSize {
			return notes, nil
		}

		currentOffset += len(page)
	}
}

func fetchNotesPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]noteListItem, error) {
	var resp notesListResponse
	if err := apiClient.GraphQL(ctx, notesQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
	}, &resp); err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}

	return displayNotes(resp.Data.Notes), nil
}

func runNotesBoolMutation(ctx context.Context, apiClient *client.Client, mutation string, field string, query string) error {
	var resp noteMutationResponse
	if err := apiClient.GraphQL(ctx, mutation, map[string]any{
		"query": query,
	}, &resp); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}

	var ok bool
	switch field {
	case "trashNotes":
		ok = resp.Data.TrashNotes
	case "restoreNotes":
		ok = resp.Data.RestoreNotes
	case "deleteNotes":
		ok = resp.Data.DeleteNotes
	default:
		return fmt.Errorf("%s: unsupported mutation field", field)
	}
	if !ok {
		return fmt.Errorf("%s: %w", field, errors.New("mutation returned false"))
	}

	return nil
}

func randomUUID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}

	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	var dst [36]byte
	hex.Encode(dst[0:8], raw[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], raw[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], raw[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], raw[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], raw[10:16])

	return string(dst[:]), nil
}

func displayNotes(notes []noteListRecord) []noteListItem {
	display := make([]noteListItem, 0, len(notes))
	for _, note := range notes {
		display = append(display, noteListItem{
			ID:        note.ID,
			Title:     note.Title,
			DeletedAt: note.DeletedAt,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
			Tags:      tagNames(note.Tags),
		})
	}

	return display
}

func tagNames(tags []api.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}

	return names
}
