package cli

import (
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// getCloudflareClient initializes the Cloudflare API client using the token from Viper.
func getCloudflareClient() (*cloudflare.API, error) {
	token := viper.GetString("api_token")
	if token == "" {
		return nil, fmt.Errorf("CLOUDFLARE_API_TOKEN environment variable is required")
	}

	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudflare client: %w", err)
	}

	return api, nil
}

// printOutput formats and prints structured data based on the requested format.
func printOutput(data interface{}, format string) error {
	switch format {
	case "json":
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Print(string(b))
	default:
		// Default to JSON if format is unsupported or text (for complex structs)
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	}
	return nil
}
