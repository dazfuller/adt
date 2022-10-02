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
	fmt.Println("Azure Digital Twin - Model CLI utility.")
	fmt.Println("Provides methods for working with the Azure Digital Twin management plane for carrying out common activities with models")
	fmt.Println("This is intended to provide simpler solutions to existing features in the SDK or existing Azure CLI")
	fmt.Println("(such as model sorting to ensure that they are uploaded/deleted in dependency order)")
	fmt.Println()
	fmt.Println("List of commands:")
	fmt.Println("  clear")
	fmt.Println("        Removes all models from the Azure Digital Twin instance")
	fmt.Println("  download")
	fmt.Println("        Downloads all models from the Azure Digital Twin instance and structures them in the output location based on their model id")
	fmt.Println("  list")
	fmt.Println("        Lists the model ids currently deployed to the Azure Digital Twin instance")
	fmt.Println("  upload")
	fmt.Println("        Uploads a set of models from local storage to the Azure Digital Twin instance")
	fmt.Println()
	os.Exit(0)
}

func main() {
	var adtEndpoint string
	var useAzureCliCredentials bool
	var tenantId string
	var clientId string
	var clientSecret string
	var verbose bool
	var source cli.ModelDirectory
	var fileExtension string

	var selectedFlagSet *flag.FlagSet = nil

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	clearCommand := flag.NewFlagSet("clear", flag.ExitOnError)
	uploadCommand := flag.NewFlagSet("upload", flag.ExitOnError)
	downloadCommand := flag.NewFlagSet("download", flag.ExitOnError)

	uploadCommand.Var(&source, "source", "Directory containing the model files to upload")
	downloadCommand.Var(&source, "output", "Directory to write models to during download")
	downloadCommand.StringVar(&fileExtension, "ext", "dtdl", "File extension to use for files downloaded (valid values are 'dtdl' or 'json')")

	// Set up common flags
	for _, fs := range []*flag.FlagSet{listCommand, clearCommand, uploadCommand, downloadCommand} {
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
	case "upload":
		if len(os.Args) < 5 {
			uploadCommand.Usage()
			os.Exit(-1)
		}
		_ = uploadCommand.Parse(os.Args[2:])
		selectedFlagSet = uploadCommand
	case "download":
		if len(os.Args) < 5 {
			downloadCommand.Usage()
			os.Exit(-1)
		}
		if strings.ToLower(fileExtension) != "dtdl" && strings.ToLower(fileExtension) != "json" {
			downloadCommand.Usage()
			os.Exit(-1)
		}
		_ = downloadCommand.Parse(os.Args[2:])
		selectedFlagSet = downloadCommand
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
	} else if uploadCommand.Parsed() {
		err := cli.UploadModels(adtEndpoint, authenticationMethod, source)
		if err != nil {
			fmt.Println(err)
			os.Exit(-2)
		}
	} else if downloadCommand.Parsed() {
		err := cli.DownloadModels(adtEndpoint, authenticationMethod, source, fileExtension)
		if err != nil {
			fmt.Println(err)
			os.Exit(-2)
		}
	}
}
