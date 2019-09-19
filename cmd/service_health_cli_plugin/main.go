package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

// BasicPlugin is the struct implementing the interface defined by the core CLI. It can
// be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type ServiceHealthPlugin struct{}

// Run must be implemented by any plugin because it is part of the
// plugin interface defined by the core CLI.
//
// Run(....) is the entry point when the core CLI is invoking a command defined
// by a plugin. The first parameter, plugin.CliConnection, is a struct that can
// be used to invoke cli commands. The second paramter, args, is a slice of
// strings. args[0] will be the name of the command, and will be followed by
// any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing
// user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
// 1 should the plugin exits nonzero.
func (c *ServiceHealthPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Ensure that we called the command basic-plugin-command
	if args[0] != "service-health" || len(args) != 2 {
		fmt.Println(c.GetMetadata().Commands[0].UsageDetails.Usage)
		os.Exit(1)
	}

	// Now we know the right CLI plugin was invoked, and that we have
	// the name of a service instance to query.
	serviceInstance := args[1]

	// TODO include org, space, and username per https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide
	// "Getting health for service instance <b>mySQL in org <b>system / space <b>system as <b>admin..."
	fmt.Printf("Getting health for service instance %s...\n", serviceInstance)

	serviceModel, err := cliConnection.GetService(serviceInstance)
	if err != nil {
		fmt.Printf("FAILED\n\nCould not find service instance %s\n", serviceInstance)
		os.Exit(1)
	}
	apiEndpoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		fmt.Printf("FAILED\n\nCould not retrieve api endpoint, please use `cf api` to set it\n")
		os.Exit(1)
	}
	registryEndpoint := strings.Replace(apiEndpoint, "api.sys.", "indicator-protocol-acceptance-proxy.apps.", 1)

	token, err := cliConnection.AccessToken()
	if err != nil {
		fmt.Printf("FAILED\n\nCould not generate oauth token, use `cf login` or `cf auth` to authorize\n")
		os.Exit(1)
	}

	split := strings.Split(token, " ")
	url := fmt.Sprintf("%s/v1/indicator-documents?token=%s&source_id=%s", registryEndpoint, split[1], serviceModel.Guid)
	request, _ := http.NewRequest(http.MethodGet, url, nil)

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("FAILED\n\nCould not access indicator registry: %s", err)
		os.Exit(1)
	}

	if response.StatusCode != http.StatusOK {
		fmt.Printf("FAILED\n\nCould not access indicator registry: received status code %d\n", response.StatusCode)
		os.Exit(1)
	}

	indiDocs := make([]v1.IndicatorDocument, 0)
	err = json.NewDecoder(response.Body).Decode(&indiDocs)
	if err != nil {
		fmt.Printf("FAILED\n\nCould not decode indicator documents: %s", err)
		os.Exit(1)
	}

	table := terminal.NewTable([]string{"indicator", "description"})

	for _, doc := range indiDocs {
		for _, indicator := range doc.Spec.Indicators {
			var description string

			docDescription, ok := indicator.Documentation["description"]
			if ok {
				description = docDescription
			}

			table.Add(indicator.Name, description)
		}
	}
	fmt.Println("OK")
	fmt.Println()
	_ = table.PrintTo(os.Stdout)
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as seperate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf basic-plugin-command` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.
func (c *ServiceHealthPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "ServiceHealthPlugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "service-health",
				HelpText: "Display health indicators of a service instance if available.",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "service-health\n   cf service-health SERVICE_INSTANCE",
				},
			},
		},
	}
}

// Unlike most Go programs, the `Main()` function will not be used to run all of the
// commands provided in your plugin. Main will be used to initialize the plugin
// process, as well as any dependencies you might require for your
// plugin.
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(ServiceHealthPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
