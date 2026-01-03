package cli

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/firasmosbahi/container-composer/core"
	"gopkg.in/yaml.v3"
)

func addNetwork(composeFile *core.ComposeFile, composePath string) error {
	fmt.Println("\n🌐 Add Network Wizard\n")

	// Step 1: Network Name
	var networkName string
	if err := survey.AskOne(&survey.Input{
		Message: "Network name:",
		Help:    "Unique identifier for this network",
	}, &networkName, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	networkName = strings.TrimSpace(networkName)

	// Check conflict
	if composeFile.NetworkExists(networkName) {
		var overwrite bool
		if err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("⚠️  Network '%s' already exists. Overwrite?", networkName),
			Default: false,
		}, &overwrite); err != nil {
			return err
		}
		if !overwrite {
			return fmt.Errorf("network already exists and overwrite was declined")
		}
	}

	network := core.Network{}

	// Step 2: Driver selection
	var driver string
	if err := survey.AskOne(&survey.Select{
		Message: "Network driver:",
		Options: []string{"bridge", "host", "overlay", "macvlan", "none"},
		Default: "bridge",
		Help:    "bridge: Default Docker network driver\nhost: Use host's network stack\noverlay: Multi-host networking\nmacvlan: Assign MAC address to container\nnone: Disable networking",
	}, &driver); err != nil {
		return err
	}
	network.Driver = driver

	// Step 3: External network?
	var external bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Is this an external network?",
		Default: false,
		Help:    "External networks are managed outside of this compose file",
	}, &external); err != nil {
		return err
	}
	network.External = external

	// Preview and confirm
	if err := previewAndConfirmNetwork(composeFile, networkName, network, composePath); err != nil {
		return err
	}

	return nil
}

func previewAndConfirmNetwork(composeFile *core.ComposeFile, networkName string, network core.Network, composePath string) error {
	// Add network to compose file (temporary for preview)
	composeFile.AddNetwork(networkName, network)

	// Marshal to YAML for preview
	yamlData, err := yaml.Marshal(composeFile)
	if err != nil {
		return fmt.Errorf("failed to generate preview: %w", err)
	}

	fmt.Println("\n📄 Preview of docker-compose.yml:\n")
	fmt.Println(string(yamlData))

	var confirmed bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Apply these changes?",
		Default: true,
	}, &confirmed); err != nil {
		return err
	}

	if !confirmed {
		return fmt.Errorf("operation cancelled by user")
	}

	// Write to file
	if err := composeFile.WriteComposeFile(composePath); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	fmt.Println("\n✅ Network added successfully!")
	fmt.Printf("   Network '%s' has been added to docker-compose.yml\n", networkName)
	return nil
}
