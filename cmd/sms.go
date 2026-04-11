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
	smsPageSize = 100

	smsQuery = `query sms($offset: Int!, $limit: Int!, $query: String!) {
  sms(offset: $offset, limit: $limit, query: $query) {
    id
    body
    address
    serviceCenter
    date
    type
    threadId
    subscriptionId
    isMms
    attachments {
      path
      contentType
      name
    }
    tags {
      id
      name
      count
    }
  }
}`

	smsConversationsQuery = `query smsConversations($offset: Int!, $limit: Int!, $query: String!) {
  smsConversations(offset: $offset, limit: $limit, query: $query) {
    id
    address
    snippet
    date
    messageCount
    read
  }
}`

	sendSMSMutation = `mutation sendSms($number: String!, $body: String!) {
  sendSms(number: $number, body: $body)
}`

	sendMMSMutation = `mutation sendMms($number: String!, $body: String!, $attachmentPaths: [String!]!, $threadId: String!) {
  sendMms(number: $number, body: $body, attachmentPaths: $attachmentPaths, threadId: $threadId)
}`
)

type SMSCmd struct {
	LS            SMSLSCmd            `cmd:"" help:"List SMS messages."`
	Conversations SMSConversationsCmd `cmd:"" help:"List SMS conversations."`
	Send          SMSSendCmd          `cmd:"" help:"Send an SMS message."`
	SendMMS       SMSSendMMSCmd       `cmd:"" help:"Send an MMS message."`
}

type SMSLSCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type SMSConversationsCmd struct {
	Query  string `help:"Search query."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type SMSSendCmd struct {
	Number string `arg:"" help:"Phone number."`
	Body   string `arg:"" help:"Message body."`
}

type SMSSendMMSCmd struct {
	Number      string   `arg:"" help:"Phone number."`
	Body        string   `arg:"" help:"Message body."`
	ThreadID    string   `name:"thread-id" help:"Conversation thread ID." required:""`
	Attachments []string `help:"Attachment paths."`
}

type smsListResponse struct {
	Data struct {
		SMS []api.Message `json:"sms"`
	} `json:"data"`
}

type smsConversationsResponse struct {
	Data struct {
		SMSConversations []api.MessageConversation `json:"smsConversations"`
	} `json:"data"`
}

type smsMutationResponse struct {
	Data struct {
		SendSMS bool `json:"sendSms"`
		SendMMS bool `json:"sendMms"`
	} `json:"data"`
}

func (c *SMSLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	messages, err := listSMS(context.Background(), apiClient, c.Query, c.Offset, c.Limit)
	if err != nil {
		return err
	}

	return printer.PrintList(messages)
}

func (c *SMSConversationsCmd) Run(apiClient *client.Client, printer output.Printer) error {
	conversations, err := listSMSConversations(context.Background(), apiClient, c.Query, c.Offset, c.Limit)
	if err != nil {
		return err
	}

	return printer.PrintList(conversations)
}

func (c *SMSSendCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp smsMutationResponse
	if err := apiClient.GraphQL(context.Background(), sendSMSMutation, map[string]any{
		"body":   c.Body,
		"number": c.Number,
	}, &resp); err != nil {
		return fmt.Errorf("send sms: %w", err)
	}
	if !resp.Data.SendSMS {
		return errors.New("send sms: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("SMS sent to %s.", c.Number),
	})
}

func (c *SMSSendMMSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	attachments := c.Attachments
	if attachments == nil {
		attachments = []string{}
	}

	var resp smsMutationResponse
	if err := apiClient.GraphQL(context.Background(), sendMMSMutation, map[string]any{
		"attachmentPaths": attachments,
		"body":            c.Body,
		"number":          c.Number,
		"threadId":        c.ThreadID,
	}, &resp); err != nil {
		return fmt.Errorf("send mms: %w", err)
	}
	if !resp.Data.SendMMS {
		return errors.New("send mms: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("MMS sent to %s.", c.Number),
	})
}

func listSMS(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.Message, error) {
	if limit > 0 {
		return fetchSMSPage(ctx, apiClient, query, offset, limit)
	}

	messages := make([]api.Message, 0, smsPageSize)
	currentOffset := offset
	for {
		page, err := fetchSMSPage(ctx, apiClient, query, currentOffset, smsPageSize)
		if err != nil {
			return nil, err
		}

		messages = append(messages, page...)
		if len(page) < smsPageSize {
			return messages, nil
		}

		currentOffset += len(page)
	}
}

func fetchSMSPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.Message, error) {
	var resp smsListResponse
	if err := apiClient.GraphQL(ctx, smsQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
	}, &resp); err != nil {
		return nil, fmt.Errorf("query sms: %w", err)
	}

	return resp.Data.SMS, nil
}

func listSMSConversations(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.MessageConversation, error) {
	if limit > 0 {
		return fetchSMSConversationsPage(ctx, apiClient, query, offset, limit)
	}

	conversations := make([]api.MessageConversation, 0, smsPageSize)
	currentOffset := offset
	for {
		page, err := fetchSMSConversationsPage(ctx, apiClient, query, currentOffset, smsPageSize)
		if err != nil {
			return nil, err
		}

		conversations = append(conversations, page...)
		if len(page) < smsPageSize {
			return conversations, nil
		}

		currentOffset += len(page)
	}
}

func fetchSMSConversationsPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	offset int,
	limit int,
) ([]api.MessageConversation, error) {
	var resp smsConversationsResponse
	if err := apiClient.GraphQL(ctx, smsConversationsQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
	}, &resp); err != nil {
		return nil, fmt.Errorf("query sms conversations: %w", err)
	}

	return resp.Data.SMSConversations, nil
}
