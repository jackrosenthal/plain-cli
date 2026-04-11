package cmd

type ClipboardCmd struct {
	Set ClipboardSetCmd `cmd:"" help:"Set the clipboard text."`
}

type ClipboardSetCmd struct {
	Text string `arg:"" help:"Clipboard text."`
}
