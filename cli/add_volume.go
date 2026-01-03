package cli

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/firasmosbahi/container-composer/core"
	"gopkg.in/yaml.v3"
)

func addVolume(composeFile *core.ComposeFile, composePath string) error {
	fmt.Println("\n💾 Add Volume Wizard\n")

	// Step 1: Volume Name
	var volumeName string
	if err := survey.AskOne(&survey.Input{
		Message: "Volume name:",
		Help:    "Unique identifier for this volume",
	}, &volumeName, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	volumeName = strings.TrimSpace(volumeName)

	// Check conflict
	if composeFile.VolumeExists(volumeName) {
		var overwrite bool
		if err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("⚠️  Volume '%s' already exists. Overwrite?", volumeName),
			Default: false,
		}, &overwrite); err != nil {
			return err
		}
		if !overwrite {
			return fmt.Errorf("volume already exists and overwrite was declined")
		}
	}

	volume := core.Volume{}

	// Step 2: Driver selection
	var driver string
	if err := survey.AskOne(&survey.Select{
		Message: "Volume driver:",
		Options: []string{"local", "nfs", "custom"},
		Default: "local",
		Help:    "local: Default Docker volume driver\nnfs: Network File System\ncustom: Enter custom driver name",
	}, &driver); err != nil {
		return err
	}

	if driver == "custom" {
		var customDriver string
		if err := survey.AskOne(&survey.Input{
			Message: "Custom driver name:",
		}, &customDriver); err != nil {
			return err
		}
		volume.Driver = strings.TrimSpace(customDriver)
	} else {
		volume.Driver = driver
	}

	// Step 3: External volume?
	var external bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Is this an external volume?",
		Default: false,
		Help:    "External volumes are managed outside of this compose file",
	}, &external); err != nil {
		return err
	}
	volume.External = external

	// Preview and confirm
	if err := previewAndConfirmVolume(composeFile, volumeName, volume, composePath); err != nil {
		return err
	}

	return nil
}

func previewAndConfirmVolume(composeFile *core.ComposeFile, volumeName string, volume core.Volume, composePath string) error {
	// Add volume to compose file (temporary for preview)
	composeFile.AddVolume(volumeName, volume)

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

	fmt.Println("\n✅ Volume added successfully!")
	fmt.Printf("   Volume '%s' has been added to docker-compose.yml\n", volumeName)
	return nil
}
