package main

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/spf13/pflag"
)

var (
	resourceGroup      = pflag.StringP("resource-group", "g", "", "resource group")
	name               = pflag.StringP("name", "n", "", "image")
	source             = pflag.StringP("source", "", "", "source")
	osType             = pflag.StringP("os-type", "", "", "os-type")
	storageAccountType = pflag.StringP("storage-account-type", "", "", "storage-account-type")
)

// run does the same as `az image create -g $RESOURCEGROUP -n $IMAGE --source
// https://$STORAGEACCOUNT.blob.core.windows.net/$CONTAINER/$IMAGE.vhd --os-type
// $OSTYPE` but adds an additional argument `--storage-account-type`.
// `az image create` doesn't appear to allow controlling the SLA of the
// underlying disk.
func run() (err error) {
	ctx := context.Background()

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return err
	}

	rcli := resources.NewGroupsClient(subscriptionID)
	rcli.Authorizer = authorizer
	icli := compute.NewImagesClient(subscriptionID)
	icli.Authorizer = authorizer

	group, err := rcli.Get(ctx, *resourceGroup)
	if err != nil {
		return err
	}

	future, err := icli.CreateOrUpdate(ctx, *resourceGroup, *name, compute.Image{
		ImageProperties: &compute.ImageProperties{
			StorageProfile: &compute.ImageStorageProfile{
				OsDisk: &compute.ImageOSDisk{
					OsType:             compute.OperatingSystemTypes(*osType),
					BlobURI:            source,
					StorageAccountType: compute.StorageAccountTypes(*storageAccountType),
				},
			},
		},
		Location: group.Location,
	})
	if err != nil {
		return err
	}

	return future.WaitForCompletion(ctx, icli.Client)
}

func main() {
	pflag.Parse()

	if err := run(); err != nil {
		panic(err)
	}
}
