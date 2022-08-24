package main

import (
	"flag"
	"fmt"
	"github.com/dazfuller/adt/cli"
	"io"
	"log"
	"os"
	"strings"
)

type directory struct {
	Path      string
	Validated bool
}

func (d *directory) String() string {
	return d.Path
}

func (d *directory) Set(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("a path must be specified")
	}

	folderInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("the specified path does not exist")
	} else if err != nil {
		return fmt.Errorf("an error occured validating the path")
	} else if !folderInfo.IsDir() {
		return fmt.Errorf("the path specified is not a directory")
	}

	*d = directory{Path: path, Validated: true}
	return nil
}

func validateCredentials(adtEndpoint string, useAzureCliCredentials bool, tenantId string, clientId string, clientSecret string) (*cli.AuthenticationMethod, error) {
	if len(adtEndpoint) == 0 {
		return nil, fmt.Errorf("the Azure Digital Twin endpoint must be set")
	}

	if !strings.HasPrefix(adtEndpoint, "https://") {
		return nil, fmt.Errorf("the endpoint should start with https://")
	}

	if !useAzureCliCredentials && (len(tenantId) == 0 || len(clientId) == 0 || len(clientSecret) == 0) {
		return nil, fmt.Errorf("when not using Azure CLI credentials for access then the tenant, client id, and client secret must be specified")
	}

	var method cli.AuthenticationMethod
	if useAzureCliCredentials {
		method = cli.AuthenticationMethod{
			UseAzureCli: true,
		}
	} else {
		method = cli.AuthenticationMethod{
			UseAzureCli:  false,
			TenantId:     tenantId,
			ClientId:     clientId,
			ClientSecret: clientSecret,
		}
	}

	return &method, nil
}

func highLevelUsageAndExit() {
	fmt.Println("Expected list, clear, or upload commands to be provided")
	os.Exit(0)
}

func main() {
	var adtEndpoint string
	//var sourceDirectory directory
	var useAzureCliCredentials bool
	var tenantId string
	var clientId string
	var clientSecret string
	var verbose bool

	var selectedFlagSet *flag.FlagSet = nil

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	clearCommand := flag.NewFlagSet("clear", flag.ExitOnError)

	// Set up common flags
	for _, fs := range []*flag.FlagSet{listCommand, clearCommand} {
		fs.StringVar(&adtEndpoint, "endpoint", "", "Endpoint of the Azure digital twin instance (e.g. https://my-twin.api.weu.digitaltwins.azure.net)")
		fs.BoolVar(&useAzureCliCredentials, "use-cli", false, "Indicates if the credentials of the Azure CLI should be used")
		fs.StringVar(&tenantId, "tenant", "", "ID of the tenant to authenticate the client credentials against")
		fs.StringVar(&clientId, "client-id", "", "ID (app id) of the app registration being used for authentication")
		fs.StringVar(&clientSecret, "client-secret", "", "Secret for the app registration being used for authentication")
		fs.BoolVar(&verbose, "verbose", false, "Indicates if logging output should be displayed")
	}

	if len(os.Args) < 2 {
		highLevelUsageAndExit()
	}

	switch strings.ToLower(os.Args[1]) {
	case "list":
		if len(os.Args) < 4 {
			listCommand.Usage()
			os.Exit(-1)
		}
		_ = listCommand.Parse(os.Args[2:])
		selectedFlagSet = listCommand
	case "clear":
		if len(os.Args) < 4 {
			clearCommand.Usage()
			os.Exit(-1)
		}
		_ = clearCommand.Parse(os.Args[2:])
		selectedFlagSet = clearCommand
	default:
		highLevelUsageAndExit()
	}

	authenticationMethod, err := validateCredentials(adtEndpoint, useAzureCliCredentials, tenantId, clientId, clientSecret)
	if err != nil {
		fmt.Printf("An error occured parsing the arguments: %s\n", err)
		selectedFlagSet.Usage()
		os.Exit(-1)
	}

	if !verbose {
		log.SetOutput(io.Discard)
	}

	if listCommand.Parsed() {
		err := cli.ListModels(adtEndpoint, authenticationMethod)
		if err != nil {
			fmt.Println(err)
			os.Exit(-2)
		}
	} else if clearCommand.Parsed() {
		err := cli.ClearModels(adtEndpoint, authenticationMethod)
		if err != nil {
			fmt.Println(err)
			os.Exit(-2)
		}
	}
}
