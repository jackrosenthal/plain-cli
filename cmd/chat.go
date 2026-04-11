package cmd

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
