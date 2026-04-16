package chat

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

type Cmd struct {
	Channels ChannelsCmd `cmd:"" help:"Manage chat channels."`
	Messages MessagesCmd `cmd:"" help:"List messages for a channel."`
	Send     SendCmd     `cmd:"" help:"Send a chat message."`
	Delete   DeleteCmd   `cmd:"" help:"Delete a chat message."`
}

type ChannelsCmd struct {
	LS      ChannelsLSCmd      `cmd:"" help:"List chat channels."`
	Create  ChannelsCreateCmd  `cmd:"" help:"Create a chat channel."`
	Update  ChannelsUpdateCmd  `cmd:"" help:"Update a chat channel."`
	Delete  ChannelsDeleteCmd  `cmd:"" help:"Delete a chat channel."`
	Leave   ChannelsLeaveCmd   `cmd:"" help:"Leave a chat channel."`
	Invite  ChannelsInviteCmd  `cmd:"" help:"Respond to chat invitations."`
	Members ChannelsMembersCmd `cmd:"" help:"Manage channel members."`
}

type ChannelsLSCmd struct{}

type ChannelsCreateCmd struct {
	Name string `arg:"" help:"Channel name."`
}

type ChannelsUpdateCmd struct {
	ID   string `arg:"" help:"Channel ID."`
	Name string `arg:"" help:"New channel name."`
}

type ChannelsDeleteCmd struct {
	ID string `arg:"" help:"Channel ID."`
}

type ChannelsLeaveCmd struct {
	ID string `arg:"" help:"Channel ID."`
}

type ChannelsInviteCmd struct {
	Accept  ChannelsInviteAcceptCmd  `cmd:"" help:"Accept a channel invite."`
	Decline ChannelsInviteDeclineCmd `cmd:"" help:"Decline a channel invite."`
}

type ChannelsInviteAcceptCmd struct {
	ID string `arg:"" help:"Invitation ID."`
}

type ChannelsInviteDeclineCmd struct {
	ID string `arg:"" help:"Invitation ID."`
}

type ChannelsMembersCmd struct {
	Add    ChannelsMembersAddCmd    `cmd:"" help:"Add a channel member."`
	Remove ChannelsMembersRemoveCmd `cmd:"" help:"Remove a channel member."`
}

type ChannelsMembersAddCmd struct {
	ChannelID string `arg:"" help:"Channel ID."`
	PeerID    string `arg:"" help:"Peer ID."`
}

type ChannelsMembersRemoveCmd struct {
	ChannelID string `arg:"" help:"Channel ID."`
	PeerID    string `arg:"" help:"Peer ID."`
}

type MessagesCmd struct {
	ChannelID string `arg:"" help:"Channel ID."`
}

type SendCmd struct {
	ToID    string `arg:"" help:"Recipient or channel ID."`
	Content string `arg:"" help:"Message content."`
}

type DeleteCmd struct {
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

func (c *ChannelsLSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelsResponse
	if err := apiClient.GraphQL(context.Background(), chatChannelsQuery, nil, &resp); err != nil {
		return fmt.Errorf("query chat channels: %w", err)
	}

	return printer.PrintList(resp.Data.ChatChannels)
}

func (c *ChannelsCreateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), createChatChannelMutation, map[string]any{
		"name": c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("create chat channel: %w", err)
	}

	return printer.Print(resp.Data.CreateChatChannel)
}

func (c *ChannelsUpdateCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), updateChatChannelMutation, map[string]any{
		"id":   c.ID,
		"name": c.Name,
	}, &resp); err != nil {
		return fmt.Errorf("update chat channel: %w", err)
	}

	return printer.Print(resp.Data.UpdateChatChannel)
}

func (c *ChannelsDeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteChatChannelMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("delete chat channel: %w", err)
	}
	if !resp.Data.DeleteChatChannel {
		return errors.New("delete chat channel: mutation returned false")
	}

	return nil
}

func (c *ChannelsLeaveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), leaveChatChannelMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("leave chat channel: %w", err)
	}
	if !resp.Data.LeaveChatChannel {
		return errors.New("leave chat channel: mutation returned false")
	}

	return nil
}

func (c *ChannelsInviteAcceptCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), acceptChatChannelInviteMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("accept chat channel invite: %w", err)
	}
	if !resp.Data.AcceptChatChannelInvite {
		return errors.New("accept chat channel invite: mutation returned false")
	}

	return nil
}

func (c *ChannelsInviteDeclineCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), declineChatChannelInviteMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("decline chat channel invite: %w", err)
	}
	if !resp.Data.DeclineChatChannelInvite {
		return errors.New("decline chat channel invite: mutation returned false")
	}

	return nil
}

func (c *ChannelsMembersAddCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), addChatChannelMemberMutation, map[string]any{
		"id":     c.ChannelID,
		"peerId": c.PeerID,
	}, &resp); err != nil {
		return fmt.Errorf("add chat channel member: %w", err)
	}

	return printer.Print(resp.Data.AddChatChannelMember)
}

func (c *ChannelsMembersRemoveCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatChannelMutationResponse
	if err := apiClient.GraphQL(context.Background(), removeChatChannelMemberMutation, map[string]any{
		"id":     c.ChannelID,
		"peerId": c.PeerID,
	}, &resp); err != nil {
		return fmt.Errorf("remove chat channel member: %w", err)
	}

	return printer.Print(resp.Data.RemoveChatChannelMember)
}

func (c *MessagesCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatItemsResponse
	if err := apiClient.GraphQL(context.Background(), chatItemsQuery, map[string]any{
		"id": c.ChannelID,
	}, &resp); err != nil {
		return fmt.Errorf("query chat messages: %w", err)
	}

	return printer.PrintList(resp.Data.ChatItems)
}

func (c *SendCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatItemMutationResponse
	if err := apiClient.GraphQL(context.Background(), sendChatItemMutation, map[string]any{
		"content": c.Content,
		"toId":    c.ToID,
	}, &resp); err != nil {
		return fmt.Errorf("send chat message: %w", err)
	}

	return printer.Print(resp.Data.SendChatItem)
}

func (c *DeleteCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp chatItemMutationResponse
	if err := apiClient.GraphQL(context.Background(), deleteChatItemMutation, map[string]any{
		"id": c.ID,
	}, &resp); err != nil {
		return fmt.Errorf("delete chat message: %w", err)
	}
	if !resp.Data.DeleteChatItem {
		return errors.New("delete chat message: mutation returned false")
	}

	return nil
}
