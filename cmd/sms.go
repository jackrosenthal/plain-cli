package cmd

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
