package notifications

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	notificationsQuery = `query {
  notifications {
    id
    onlyOnce
    isClearable
    appId
    appName
    time
    silent
    title
    body
    actions
    replyActions
  }
}`

	cancelNotificationsMutation = `mutation cancelNotifications($ids: [ID!]!) {
  cancelNotifications(ids: $ids)
}`

	replyNotificationMutation = `mutation replyNotification($id: ID!, $actionIndex: Int!, $text: String!) {
  replyNotification(id: $id, actionIndex: $actionIndex, text: $text)
}`
)

type Cmd struct {
	LS     LSCmd     `cmd:"" help:"List notifications."`
	Cancel CancelCmd `cmd:"" help:"Cancel notifications."`
	Reply  ReplyCmd  `cmd:"" help:"Reply to a notification."`
}

type LSCmd struct{}

type CancelCmd struct {
	IDs []string `arg:"" name:"ids" help:"Notification IDs."`
}

type ReplyCmd struct {
	ID          string `arg:"" help:"Notification ID."`
	ActionIndex int    `name:"action-index" help:"Action index to invoke." required:""`
	Text        string `name:"text" help:"Reply text." required:""`
}

type notificationsResponse struct {
	Data struct {
		Notifications []api.Notification `json:"notifications"`
	} `json:"data"`
}

type notificationMutationResponse struct {
	Data struct {
		CancelNotifications bool `json:"cancelNotifications"`
		ReplyNotification   bool `json:"replyNotification"`
	} `json:"data"`
}

func (c *LSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp notificationsResponse
	if err := apiClient.GraphQL(context.Background(), notificationsQuery, nil, &resp); err != nil {
		return fmt.Errorf("query notifications: %w", err)
	}

	return printer.PrintList(resp.Data.Notifications)
}

func (c *CancelCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp notificationMutationResponse
	if err := apiClient.GraphQL(context.Background(), cancelNotificationsMutation, map[string]any{
		"ids": c.IDs,
	}, &resp); err != nil {
		return fmt.Errorf("cancel notifications: %w", err)
	}
	if !resp.Data.CancelNotifications {
		return errors.New("cancel notifications: mutation returned false")
	}

	return nil
}

func (c *ReplyCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp notificationMutationResponse
	if err := apiClient.GraphQL(context.Background(), replyNotificationMutation, map[string]any{
		"actionIndex": c.ActionIndex,
		"id":          c.ID,
		"text":        c.Text,
	}, &resp); err != nil {
		return fmt.Errorf("reply to notification: %w", err)
	}
	if !resp.Data.ReplyNotification {
		return errors.New("reply to notification: mutation returned false")
	}

	return nil
}
