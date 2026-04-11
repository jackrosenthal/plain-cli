package cmd

type NotificationsCmd struct {
	LS     NotificationsLSCmd     `cmd:"" help:"List notifications."`
	Cancel NotificationsCancelCmd `cmd:"" help:"Cancel notifications."`
	Reply  NotificationsReplyCmd  `cmd:"" help:"Reply to a notification."`
}

type NotificationsLSCmd struct{}

type NotificationsCancelCmd struct {
	IDs []string `arg:"" name:"ids" help:"Notification IDs."`
}

type NotificationsReplyCmd struct {
	ID          string `arg:"" help:"Notification ID."`
	ActionIndex int    `name:"action-index" help:"Action index to invoke." required:""`
	Text        string `help:"Reply text." required:""`
}
