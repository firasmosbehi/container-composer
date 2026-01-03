package cli

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/firasmosbahi/container-composer/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a service, network, or volume to docker-compose.yml",
	Long: `Interactive wizard to add resources to your docker-compose.yml file.

Supports adding:
  - Services (containers)
  - Networks
  - Volumes

The wizard will guide you through all configuration options and show a preview before applying changes.`,
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	fmt.Println("\n🚀 Add Resource Wizard\n")

	// 1. Check if docker-compose.yml exists
	composePath := "docker-compose.yml"
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found in current directory")
	}

	// 2. Parse existing docker-compose.yml
	composeFile, err := core.ParseComposeFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose.yml: %w", err)
	}

	// 3. Show resource type selection
	resourceType, err := selectResourceType()
	if err != nil {
		return err
	}

	// 4. Run appropriate wizard based on selection
	switch resourceType {
	case "service":
		return addService(composeFile, composePath)
	case "network":
		return addNetwork(composeFile, composePath)
	case "volume":
		return addVolume(composeFile, composePath)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

func selectResourceType() (string, error) {
	options := []string{
		"🐳 Service - Add a new container/service",
		"🌐 Network - Add a new network",
		"💾 Volume - Add a new volume",
	}

	resourceMap := map[string]string{
		options[0]: "service",
		options[1]: "network",
		options[2]: "volume",
	}

	var selected string
	prompt := &survey.Select{
		Message: "What do you want to add?",
		Options: options,
		Help:    "Select the type of resource to add to your docker-compose.yml",
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", err
	}

	return resourceMap[selected], nil
}
