package cmd

import (
	"context"
	"fmt"
	"math"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

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

const (
	deviceInfoQuery = `query {
  deviceInfo {
    deviceName
    releaseBuildVersion
    versionCodeName
    manufacturer
    securityPatch
    bootloader
    deviceId
    model
    product
    fingerprint
    hardware
    radioVersion
    device
    board
    displayVersion
    buildBrand
    buildHost
    buildTime
    uptime
    buildUser
    serial
    osVersion
    language
    sdkVersion
    javaVmVersion
    kernelVersion
    glEsVersion
    screenDensity
    screenHeight
    screenWidth
    phoneNumbers {
      id
      name
      number
    }
  }
}`

	deviceBatteryQuery = `query {
  battery {
    level
    voltage
    health
    plugged
    temperature
    status
    technology
    capacity
  }
}`

	deviceAppQuery = `query {
  app {
    usbConnected
    urlToken
    httpPort
    httpsPort
    appDir
    deviceName
    battery
    appVersion
    osVersion
    channel
    permissions
    audios {
      title
      artist
      path
      duration
    }
    audioCurrent
    audioMode
    sdcardPath
    usbDiskPaths
    internalStoragePath
    downloadsDir
    developerMode
    favoriteFolders {
      rootPath
      fullPath
      alias
    }
  }
}`

	devicePeersQuery = `query {
  peers {
    id
    name
    ip
    status
    port
    deviceType
    createdAt
    updatedAt
  }
}`

	deviceMountsQuery = `query {
  mounts {
    id
    name
    path
    mountPoint
    fsType
    totalBytes
    usedBytes
    freeBytes
    remote
    alias
    driveType
    diskID
  }
}`

	deviceRelaunchMutation = `mutation {
  relaunchApp
}`
)

type deviceInfoResponse struct {
	Data struct {
		DeviceInfo api.DeviceInfo `json:"deviceInfo"`
	} `json:"data"`
}

type deviceBatteryResponse struct {
	Data struct {
		Battery api.Battery `json:"battery"`
	} `json:"data"`
}

type deviceAppResponse struct {
	Data struct {
		App api.App `json:"app"`
	} `json:"data"`
}

type devicePeersResponse struct {
	Data struct {
		Peers []api.Peer `json:"peers"`
	} `json:"data"`
}

type deviceMountsResponse struct {
	Data struct {
		Mounts []api.Mount `json:"mounts"`
	} `json:"data"`
}

type deviceRelaunchResponse struct {
	Data struct {
		RelaunchApp bool `json:"relaunchApp"`
	} `json:"data"`
}

type mountOutput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	MountPoint string `json:"mountPoint"`
	FSType     string `json:"fsType"`
	TotalBytes string `json:"totalBytes"`
	UsedBytes  string `json:"usedBytes"`
	FreeBytes  string `json:"freeBytes"`
	Remote     bool   `json:"remote"`
	Alias      string `json:"alias"`
	DriveType  string `json:"driveType"`
	DiskID     string `json:"diskID"`
}

type deviceStatusMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (c *DeviceInfoCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp deviceInfoResponse
	if err := apiClient.GraphQL(context.Background(), deviceInfoQuery, nil, &resp); err != nil {
		return fmt.Errorf("query device info: %w", err)
	}

	return printer.Print(resp.Data.DeviceInfo)
}

func (c *DeviceBatteryCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp deviceBatteryResponse
	if err := apiClient.GraphQL(context.Background(), deviceBatteryQuery, nil, &resp); err != nil {
		return fmt.Errorf("query battery: %w", err)
	}

	return printer.Print(resp.Data.Battery)
}

func (c *DeviceAppCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp deviceAppResponse
	if err := apiClient.GraphQL(context.Background(), deviceAppQuery, nil, &resp); err != nil {
		return fmt.Errorf("query app: %w", err)
	}

	return printer.Print(resp.Data.App)
}

func (c *DevicePeersCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp devicePeersResponse
	if err := apiClient.GraphQL(context.Background(), devicePeersQuery, nil, &resp); err != nil {
		return fmt.Errorf("query peers: %w", err)
	}

	return printer.PrintList(resp.Data.Peers)
}

func (c *DeviceMountsCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp deviceMountsResponse
	if err := apiClient.GraphQL(context.Background(), deviceMountsQuery, nil, &resp); err != nil {
		return fmt.Errorf("query mounts: %w", err)
	}

	formatted := make([]mountOutput, 0, len(resp.Data.Mounts))
	for _, mount := range resp.Data.Mounts {
		formatted = append(formatted, mountOutput{
			ID:         mount.ID,
			Name:       mount.Name,
			Path:       mount.Path,
			MountPoint: mount.MountPoint,
			FSType:     mount.FSType,
			TotalBytes: humanizeBytes(mount.TotalBytes),
			UsedBytes:  humanizeBytes(mount.UsedBytes),
			FreeBytes:  humanizeBytes(mount.FreeBytes),
			Remote:     mount.Remote,
			Alias:      mount.Alias,
			DriveType:  mount.DriveType,
			DiskID:     mount.DiskID,
		})
	}

	return printer.PrintList(formatted)
}

func (c *DeviceRelaunchCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp deviceRelaunchResponse
	if err := apiClient.GraphQL(context.Background(), deviceRelaunchMutation, nil, &resp); err != nil {
		return fmt.Errorf("relaunch app: %w", err)
	}

	if !resp.Data.RelaunchApp {
		return fmt.Errorf("relaunch app: mutation returned false")
	}

	return printer.Print(deviceStatusMessage{
		Status:  "ok",
		Message: "App relaunch requested.",
	})
}

func humanizeBytes(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	value := float64(size)
	for idx, unit := range units {
		value /= 1024
		if value < 1024 || idx == len(units)-1 {
			if math.Mod(value, 1) == 0 {
				return fmt.Sprintf("%.0f %s", value, unit)
			}

			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}

	return fmt.Sprintf("%d B", size)
}
