package cmd

type TagsCmd struct {
	LS     TagsLSCmd     `cmd:"" help:"List tags for a content type."`
	Create TagsCreateCmd `cmd:"" help:"Create a tag."`
	Update TagsUpdateCmd `cmd:"" help:"Update a tag."`
	Delete TagsDeleteCmd `cmd:"" help:"Delete a tag."`
	Add    TagsAddCmd    `cmd:"" help:"Add items to tags."`
	Remove TagsRemoveCmd `cmd:"" help:"Remove items from tags."`
}

type TagsLSCmd struct {
	Type string `arg:"" help:"Content type." enum:"image,video,audio,note,feed-entry,call,contact,message,bookmark"`
}

type TagsCreateCmd struct {
	Name string `arg:"" help:"Tag name."`
}

type TagsUpdateCmd struct {
	ID   string `arg:"" help:"Tag ID."`
	Name string `arg:"" help:"New tag name."`
}

type TagsDeleteCmd struct {
	ID string `arg:"" help:"Tag ID."`
}

type TagsAddCmd struct {
	Type  string `arg:"" help:"Content type." enum:"image,video,audio,note,feed-entry,call,contact,message,bookmark"`
	Query string `arg:"" help:"Selection query."`
	Tags  string `help:"Comma-separated tag names." required:""`
}

type TagsRemoveCmd struct {
	Type  string `arg:"" help:"Content type." enum:"image,video,audio,note,feed-entry,call,contact,message,bookmark"`
	Query string `arg:"" help:"Selection query."`
}
