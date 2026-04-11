package cmd

type PackagesCmd struct {
	LS        PackagesLSCmd        `cmd:"" help:"List installed packages."`
	Install   PackagesInstallCmd   `cmd:"" help:"Install a package from the device."`
	Uninstall PackagesUninstallCmd `cmd:"" help:"Uninstall packages."`
}

type PackagesLSCmd struct {
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type PackagesInstallCmd struct {
	DevicePath string `arg:"" help:"Device path to the package file."`
}

type PackagesUninstallCmd struct {
	IDs []string `arg:"" name:"ids" help:"Package IDs to uninstall."`
}
