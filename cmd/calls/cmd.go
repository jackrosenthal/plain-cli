package calls

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/cmdutil"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	callsPageSize = 100

	callsQuery = `query calls($offset: Int!, $limit: Int!, $query: String!) {
  calls(offset: $offset, limit: $limit, query: $query) {
    id
    name
    number
    duration
    accountId
    startedAt
    photoId
    type
    geo {
      isp
      city
      province
    }
    tags {
      id
      name
      count
    }
  }
}`

	callMutation = `mutation call($number: String!) {
  call(number: $number)
}`

	deleteCallsMutation = `mutation deleteCalls($query: String!) {
  deleteCalls(query: $query)
}`
)

type Cmd struct {
	LS     LSCmd     `cmd:"" help:"List calls."`
	Call   CallCmd   `cmd:"" help:"Place a call."`
	Delete DeleteCmd `cmd:"" help:"Delete calls."`
}

type LSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type CallCmd struct {
	Number string `arg:"" help:"Phone number to call."`
}

type DeleteCmd struct {
	Query string `arg:"" help:"Selection query."`
}

type callsListResponse struct {
	Data struct {
		Calls []api.Call `json:"calls"`
	} `json:"data"`
}

type callsMutationResponse struct {
	Data struct {
		Call        bool `json:"call"`
		DeleteCalls bool `json:"deleteCalls"`
	} `json:"data"`
}

type callDisplay struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Number    string    `json:"number"`
	Duration  int       `json:"duration"`
	AccountID string    `json:"accountId"`
	StartedAt string    `json:"startedAt"`
	PhotoID   string    `json:"photoId"`
	Type      string    `json:"type"`
	Geo       api.Geo   `json:"geo"`
	Tags      []api.Tag `json:"tags"`
}

func (c *LSCmd) Run(cli *cmdutil.CLIContext, apiClient *client.Client, printer output.Printer) error {
	calls, err := listCalls(context.Background(), apiClient, c.Query, c.Offset, c.Limit)
	if err != nil {
		return err
	}

	if cli != nil && cli.Output == output.FormatJSON {
		return printer.PrintList(calls)
	}

	return printer.PrintList(displayCalls(calls))
}

func (c *CallCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp callsMutationResponse
	if err := apiClient.GraphQL(context.Background(), callMutation, map[string]any{
		"number": c.Number,
	}, &resp); err != nil {
		return fmt.Errorf("place call: %w", err)
	}
	if !resp.Data.Call {
		return errors.New("place call: mutation returned false")
	}

	return nil
}

func (c *DeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp callsMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteCallsMutation, map[string]any{
		"query": c.Query,
	}, &resp); err != nil {
		return fmt.Errorf("delete calls: %w", err)
	}
	if !resp.Data.DeleteCalls {
		return errors.New("delete calls: mutation returned false")
	}

	return nil
}

func listCalls(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.Call, error) {
	if limit > 0 {
		return fetchCallsPage(ctx, apiClient, query, offset, limit)
	}

	calls := make([]api.Call, 0, callsPageSize)
	currentOffset := offset
	for {
		page, err := fetchCallsPage(ctx, apiClient, query, currentOffset, callsPageSize)
		if err != nil {
			return nil, err
		}

		calls = append(calls, page...)
		if len(page) < callsPageSize {
			return calls, nil
		}

		currentOffset += len(page)
	}
}

func fetchCallsPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.Call, error) {
	var resp callsListResponse
	if err := apiClient.GraphQL(ctx, callsQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
	}, &resp); err != nil {
		return nil, fmt.Errorf("query calls: %w", err)
	}

	return resp.Data.Calls, nil
}

func displayCalls(calls []api.Call) []callDisplay {
	display := make([]callDisplay, 0, len(calls))
	for _, call := range calls {
		display = append(display, callDisplay{
			ID:        call.ID,
			Name:      call.Name,
			Number:    call.Number,
			Duration:  call.Duration,
			AccountID: call.AccountID,
			StartedAt: call.StartedAt,
			PhotoID:   call.PhotoID,
			Type:      callTypeLabel(call.Type),
			Geo:       call.Geo,
			Tags:      call.Tags,
		})
	}

	return display
}

func callTypeLabel(callType int) string {
	switch callType {
	case 1:
		return "incoming"
	case 2:
		return "outgoing"
	case 3:
		return "missed"
	default:
		return fmt.Sprintf("%d", callType)
	}
}
