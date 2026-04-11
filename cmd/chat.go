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
	chatChannelsQuery = `query {
  chatChannels {
    id
    name
    owner
    members {
      id
      status
    }
    version
    status
    createdAt
    updatedAt
  }
}`

	createChatChannelMutation = `mutation createChatChannel($name: String!) {
  createChatChannel(name: $name) {
    id
    name
    owner
    members {
      id
      status
    }
    version
    status
    createdAt
    updatedAt
  }
}`

	updateChatChannelMutation = `mutation updateChatChannel($id: ID!, $name: String!) {
  updateChatChannel(id: $id, name: $name) {
    id
    name
    owner
    members {
      id
      status
    }
    version
    status
    createdAt
    updatedAt
  }
}`

	deleteChatChannelMutation = `mutation deleteChatChannel($id: ID!) {
  deleteChatChannel(id: $id)
}`

	leaveChatChannelMutation = `mutation leaveChatChannel($id: ID!) {
  leaveChatChannel(id: $id)
}`

	acceptChatChannelInviteMutation = `mutation acceptChatChannelInvite($id: ID!) {
  acceptChatChannelInvite(id: $id)
}`

	declineChatChannelInviteMutation = `mutation declineChatChannelInvite($id: ID!) {
  declineChatChannelInvite(id: $id)
}`

	addChatChannelMemberMutation = `mutation addChatChannelMember($id: ID!, $peerId: String!) {
  addChatChannelMember(id: $id, peerId: $peerId) {
    id
    name
    owner
    members {
      id
      status
    }
    version
    status
    createdAt
    updatedAt
  }
}`

	removeChatChannelMemberMutation = `mutation removeChatChannelMember($id: ID!, $peerId: String!) {
  removeChatChannelMember(id: $id, peerId: $peerId) {
    id
    name
    owner
    members {
      id
      status
    }
    version
    status
    createdAt
    updatedAt
  }
}`

	chatItemsQuery = `query chatItems($id: String!) {
  chatItems(id: $id) {
    id
    fromId
    toId
    channelId
    createdAt
    content
    data
  }
}`

	sendChatItemMutation = `mutation sendChatItem($toId: String!, $content: String!) {
  sendChatItem(toId: $toId, content: $content) {
    id
    fromId
    toId
    channelId
    createdAt
    content
    data
  }
}`

	deleteChatItemMutation = `mutation deleteChatItem($id: ID!) {
  deleteChatItem(id: $id)
}`
)

type ChatCmd struct {
	Channels ChatChannelsCmd `cmd:"" help:"Manage chat channels."`
	Messages ChatMessagesCmd `cmd:"" help:"List messages for a channel."`
	Send     ChatSendCmd     `cmd:"" help:"Send a chat message."`
	Delete   ChatDeleteCmd   `cmd:"" help:"Delete a chat message."`
}

type ChatChannelsCmd struct {
	LS      ChatChannelsLSCmd      `cmd:"" help:"List chat channels."`
	Create  ChatChannelsCreateCmd  `cmd:"" help:"Create a chat channel."`
	Update  ChatChannelsUpdateCmd  `cmd:"" help:"Update a chat channel."`
	Delete  ChatChannelsDeleteCmd  `cmd:"" help:"Delete a chat channel."`
	Leave   ChatChannelsLeaveCmd   `cmd:"" help:"Leave a chat channel."`
	Invite  ChatChannelsInviteCmd  `cmd:"" help:"Respond to chat invitations."`
	Members ChatChannelsMembersCmd `cmd:"" help:"Manage channel members."`
}

type ChatChannelsLSCmd struct{}

type ChatChannelsCreateCmd struct {
	Name string `arg:"" help:"Channel name."`
}

type ChatChannelsUpdateCmd struct {
	ID   string `arg:"" help:"Channel ID."`
	Name string `arg:"" help:"New channel name."`
}

type ChatChannelsDeleteCmd struct {
	ID string `arg:"" help:"Channel ID."`
}

type ChatChannelsLeaveCmd struct {
	ID string `arg:"" help:"Channel ID."`
}

type ChatChannelsInviteCmd struct {
	Accept  ChatChannelsInviteAcceptCmd  `cmd:"" help:"Accept a channel invite."`
	Decline ChatChannelsInviteDeclineCmd `cmd:"" help:"Decline a channel invite."`
}

type ChatChannelsInviteAcceptCmd struct {
	ID string `arg:"" help:"Invitation ID."`
}

type ChatChannelsInviteDeclineCmd struct {
	ID string `arg:"" help:"Invitation ID."`
}

type ChatChannelsMembersCmd struct {
	Add    ChatChannelsMembersAddCmd    `cmd:"" help:"Add a channel member."`
	Remove ChatChannelsMembersRemoveCmd `cmd:"" help:"Remove a channel member."`
}

type ChatChannelsMembersAddCmd struct {
	ChannelID string `arg:"" help:"Channel ID."`
	PeerID    string `arg:"" help:"Peer ID."`
}

type ChatChannelsMembersRemoveCmd struct {
	ChannelID string `arg:"" help:"Channel ID."`
	PeerID    string `arg:"" help:"Peer ID."`
}

type ChatMessagesCmd struct {
	ChannelID string `arg:"" help:"Channel ID."`
}

type ChatSendCmd struct {
	ToID    string `arg:"" help:"Recipient or channel ID."`
	Content string `arg:"" help:"Message content."`
}

type ChatDeleteCmd struct {
	ID string `arg:"" help:"Chat item ID."`
}

type chatChannelsResponse struct {
	Data struct {
		ChatChannels []api.ChatChannel `json:"chatChannels"`
	} `json:"data"`
}

type chatChannelMutationResponse struct {
	Data struct {
		CreateChatChannel        api.ChatChannel `json:"createChatChannel"`
		UpdateChatChannel        api.ChatChannel `json:"updateChatChannel"`
		AddChatChannelMember     api.ChatChannel `json:"addChatChannelMember"`
		RemoveChatChannelMember  api.ChatChannel `json:"removeChatChannelMember"`
		DeleteChatChannel        bool            `json:"deleteChatChannel"`
		LeaveChatChannel         bool            `json:"leaveChatChannel"`
		AcceptChatChannelInvite  bool            `json:"acceptChatChannelInvite"`
		DeclineChatChannelInvite bool            `json:"declineChatChannelInvite"`
	} `json:"data"`
}

type chatItemsResponse struct {
	Data struct {
		ChatItems []api.ChatItem `json:"chatItems"`
	} `json:"data"`
}

type chatItemMutationResponse struct {
	Data struct {
		SendChatItem   api.ChatItem `json:"sendChatItem"`
		DeleteChatItem bool         `json:"deleteChatItem"`
	} `json:"data"`
}

func (c *ChatChannelsLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelsResponse
	if err := apiClient.GraphQL(context.Background(), chatChannelsQuery, nil, &resp); err != nil {
		return fmt.Errorf("query chat channels: %w", err)
	}

	return printer.PrintList(resp.Data.ChatChannels)
}

func (c *ChatChannelsCreateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), createChatChannelMutation, map[string]any{
		"name": c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("create chat channel: %w", err)
	}

	return printer.Print(resp.Data.CreateChatChannel)
}

func (c *ChatChannelsUpdateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateChatChannelMutation, map[string]any{
		"id":   c.ID,
		"name": c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("update chat channel: %w", err)
	}

	return printer.Print(resp.Data.UpdateChatChannel)
}

func (c *ChatChannelsDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteChatChannelMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("delete chat channel: %w", err)
	}
	if !resp.Data.DeleteChatChannel {
		return errors.New("delete chat channel: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Deleted chat channel %s.", c.ID),
	})
}

func (c *ChatChannelsLeaveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), leaveChatChannelMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("leave chat channel: %w", err)
	}
	if !resp.Data.LeaveChatChannel {
		return errors.New("leave chat channel: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Left chat channel %s.", c.ID),
	})
}

func (c *ChatChannelsInviteAcceptCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), acceptChatChannelInviteMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("accept chat channel invite: %w", err)
	}
	if !resp.Data.AcceptChatChannelInvite {
		return errors.New("accept chat channel invite: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Accepted chat channel invite %s.", c.ID),
	})
}

func (c *ChatChannelsInviteDeclineCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), declineChatChannelInviteMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("decline chat channel invite: %w", err)
	}
	if !resp.Data.DeclineChatChannelInvite {
		return errors.New("decline chat channel invite: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Declined chat channel invite %s.", c.ID),
	})
}

func (c *ChatChannelsMembersAddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), addChatChannelMemberMutation, map[string]any{
		"id":     c.ChannelID,
		"peerId": c.PeerID,
	}, &resp); err != nil {
		return fmt.Errorf("add chat channel member: %w", err)
	}

	return printer.Print(resp.Data.AddChatChannelMember)
}

func (c *ChatChannelsMembersRemoveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), removeChatChannelMemberMutation, map[string]any{
		"id":     c.ChannelID,
		"peerId": c.PeerID,
	}, &resp); err != nil {
		return fmt.Errorf("remove chat channel member: %w", err)
	}

	return printer.Print(resp.Data.RemoveChatChannelMember)
}

func (c *ChatMessagesCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatItemsResponse
	if err := apiClient.GraphQL(context.Background(), chatItemsQuery, map[string]any{
		"id": c.ChannelID,
	}, &resp); err != nil {
		return fmt.Errorf("query chat messages: %w", err)
	}

	return printer.PrintList(resp.Data.ChatItems)
}

func (c *ChatSendCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatItemMutationResponse
	if err := apiClient.GraphQL(context.Background(), sendChatItemMutation, map[string]any{
		"content": c.Content,
		"toId":    c.ToID,
	}, &resp); err != nil {
		return fmt.Errorf("send chat message: %w", err)
	}

	return printer.Print(resp.Data.SendChatItem)
}

func (c *ChatDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatItemMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteChatItemMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("delete chat message: %w", err)
	}
	if !resp.Data.DeleteChatItem {
		return errors.New("delete chat message: mutation returned false")
	}

	return printer.Print(mutationStatus{
		Status:  "ok",
		Message: fmt.Sprintf("Deleted chat message %s.", c.ID),
	})
}
