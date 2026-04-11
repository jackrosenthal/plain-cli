package cmd

type DeviceCmd struct {
	Info     DeviceInfoCmd     `cmd:"" help:"Show device information."`
	Battery  DeviceBatteryCmd  `cmd:"" help:"Show battery status."`
	App      DeviceAppCmd      `cmd:"" help:"Show app information."`
	Peers    DevicePeersCmd    `cmd:"" help:"List connected peers."`
	Mounts   DeviceMountsCmd   `cmd:"" help:"List storage mounts."`
	Relaunch DeviceRelaunchCmd `cmd:"" help:"Relaunch the Plain app."`
}

type (
	DeviceInfoCmd     struct{}
	DeviceBatteryCmd  struct{}
	DeviceAppCmd      struct{}
	DevicePeersCmd    struct{}
	DeviceMountsCmd   struct{}
	DeviceRelaunchCmd struct{}
)
