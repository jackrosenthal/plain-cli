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
	contactsPageSize = 100

	contactsQuery = `query contacts($offset: Int!, $limit: Int!, $query: String!) {
  contacts(offset: $offset, limit: $limit, query: $query) {
    id
    prefix
    suffix
    firstName
    middleName
    lastName
    updatedAt
    notes
    source
    thumbnailId
    starred
    phoneNumbers {
      label
      value
      type
      normalizedNumber
    }
    addresses {
      label
      value
      type
    }
    emails {
      label
      value
      type
    }
    websites {
      label
      value
      type
    }
    events {
      label
      value
      type
    }
    ims {
      label
      value
      type
    }
    tags {
      id
      name
      count
    }
  }
}`

	contactSourcesQuery = `query {
  contactSources {
    name
    type
  }
}`

	deleteContactsMutation = `mutation deleteContacts($query: String!) {
  deleteContacts(query: $query)
}`
)

type ContactsCmd struct {
	LS      ContactsLSCmd      `cmd:"" help:"List contacts."`
	Sources ContactsSourcesCmd `cmd:"" help:"List contact sources."`
	Delete  ContactsDeleteCmd  `cmd:"" help:"Delete contacts."`
}

type ContactsLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type ContactsSourcesCmd struct{}

type ContactsDeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type contactsListResponse struct {
	Data struct {
		Contacts []api.Contact `json:"contacts"`
	} `json:"data"`
}

type contactSourcesResponse struct {
	Data struct {
		ContactSources []contactSource `json:"contactSources"`
	} `json:"data"`
}

type deleteContactsResponse struct {
	Data struct {
		DeleteContacts bool `json:"deleteContacts"`
	} `json:"data"`
}

type contactSource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (c *ContactsLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	contacts, err := listContacts(context.Background(), apiClient, c.Query, c.Offset, c.Limit)
	if err != nil {
		return err
	}

	return printer.PrintList(contacts)
}

func (c *ContactsSourcesCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp contactSourcesResponse
	if err := apiClient.GraphQL(context.Background(), contactSourcesQuery, nil, &resp); err != nil {
		return fmt.Errorf("query contact sources: %w", err)
	}

	return printer.PrintList(resp.Data.ContactSources)
}

func (c *ContactsDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp deleteContactsResponse
	if err := apiClient.GraphQL(context.Background(), deleteContactsMutation, map[string]any{
		"query": c.Query,
	}, &resp); err != nil {
		return fmt.Errorf("delete contacts: %w", err)
	}
	if !resp.Data.DeleteContacts {
		return errors.New("delete contacts: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Deleted contacts matching %q.", c.Query),
	})
}

func listContacts(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.Contact, error) {
	if limit > 0 {
		return fetchContactsPage(ctx, apiClient, query, offset, limit)
	}

	contacts := make([]api.Contact, 0, contactsPageSize)
	currentOffset := offset
	for {
		page, err := fetchContactsPage(ctx, apiClient, query, currentOffset, contactsPageSize)
		if err != nil {
			return nil, err
		}

		contacts = append(contacts, page...)
		if len(page) < contactsPageSize {
			return contacts, nil
		}

		currentOffset += len(page)
	}
}

func fetchContactsPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.Contact, error) {
	var resp contactsListResponse
	if err := apiClient.GraphQL(ctx, contactsQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
	}, &resp); err != nil {
		return nil, fmt.Errorf("query contacts: %w", err)
	}

	return resp.Data.Contacts, nil
}
