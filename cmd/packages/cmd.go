package packages

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackrosenthal/plain-cli/internal/api"
	"github.com/jackrosenthal/plain-cli/internal/client"
	"github.com/jackrosenthal/plain-cli/internal/output"
)

const (
	packagesPageSize = 100

	packagesQuery = `query packages($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  packages(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    id
    name
    type
    version
    path
    size
    certs {
      issuer
      subject
      serialNumber
      validFrom
      validTo
    }
    installedAt
    updatedAt
  }
}`

	installPackageMutation = `mutation installPackage($path: String!) {
  installPackage(path: $path) {
    packageName
    updatedAt
    isNew
  }
}`

	uninstallPackagesMutation = `mutation uninstallPackages($ids: [ID!]!) {
  uninstallPackages(ids: $ids)
}`
)

type Cmd struct {
	LS        LSCmd        `cmd:"" help:"List installed packages."`
	Install   InstallCmd   `cmd:"" help:"Install a package from the device."`
	Uninstall UninstallCmd `cmd:"" help:"Uninstall packages."`
}

type LSCmd struct {
	Query  string `help:"Search query."`
	Sort   string `help:"Sort field."`
	Limit  int    `help:"Maximum number of results to return."`
	Offset int    `help:"Number of results to skip."`
}

type InstallCmd struct {
	DevicePath string `arg:"" help:"Device path to the package file."`
}

type UninstallCmd struct {
	IDs []string `arg:"" name:"ids" help:"Package IDs to uninstall."`
}

type packagesListResponse struct {
	Data struct {
		Packages []api.Package `json:"packages"`
	} `json:"data"`
}

type packagesMutationResponse struct {
	Data struct {
		InstallPackage    packageInstallResult `json:"installPackage"`
		UninstallPackages bool                 `json:"uninstallPackages"`
	} `json:"data"`
}

type packageInstallResult struct {
	PackageName string `json:"packageName"`
	UpdatedAt   string `json:"updatedAt"`
	IsNew       bool   `json:"isNew"`
}

func (c *LSCmd) Run(apiClient *client.Client, printer output.Printer) error {
	packages, err := listPackages(
		context.Background(),
		apiClient,
		c.Query,
		api.FileSortBy(c.Sort),
		c.Offset,
		c.Limit,
	)
	if err != nil {
		return err
	}

	return printer.PrintList(packages)
}

func (c *InstallCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp packagesMutationResponse
	if err := apiClient.GraphQL(context.Background(), installPackageMutation, map[string]any{
		"path": c.DevicePath,
	}, &resp); err != nil {
		return fmt.Errorf("install package: %w", err)
	}

	return printer.Print(resp.Data.InstallPackage)
}

func (c *UninstallCmd) Run(apiClient *client.Client, printer output.Printer) error {
	var resp packagesMutationResponse
	if err := apiClient.GraphQL(context.Background(), uninstallPackagesMutation, map[string]any{
		"ids": c.IDs,
	}, &resp); err != nil {
		return fmt.Errorf("uninstall packages: %w", err)
	}
	if !resp.Data.UninstallPackages {
		return errors.New("uninstall packages: mutation returned false")
	}

	return nil
}

func listPackages(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Package, error) {
	if sortBy == "" {
		sortBy = api.FileSortByName
	}

	if limit > 0 {
		return fetchPackagesPage(ctx, apiClient, query, sortBy, offset, limit)
	}

	packages := make([]api.Package, 0, packagesPageSize)
	currentOffset := offset
	for {
		page, err := fetchPackagesPage(ctx, apiClient, query, sortBy, currentOffset, packagesPageSize)
		if err != nil {
			return nil, err
		}

		packages = append(packages, page...)
		if len(page) < packagesPageSize {
			return packages, nil
		}

		currentOffset += len(page)
	}
}

func fetchPackagesPage(
	ctx context.Context,
	apiClient *client.Client,
	query string,
	sortBy api.FileSortBy,
	offset int,
	limit int,
) ([]api.Package, error) {
	var resp packagesListResponse
	if err := apiClient.GraphQL(ctx, packagesQuery, map[string]any{
		"limit":  limit,
		"offset": offset,
		"query":  query,
		"sortBy": sortBy.ToGraphQL(),
	}, &resp); err != nil {
		return nil, fmt.Errorf("query packages: %w", err)
	}

	return resp.Data.Packages, nil
}
