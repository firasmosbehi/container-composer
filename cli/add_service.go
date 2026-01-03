package cli

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/firasmosbahi/container-composer/core"
	"gopkg.in/yaml.v3"
)

func addService(composeFile *core.ComposeFile, composePath string) error {
	fmt.Println("\n🐳 Add Service Wizard\n")

	service := core.Service{}

	// Step 1: Service Name
	var serviceName string
	if err := survey.AskOne(&survey.Input{
		Message: "Service name:",
		Help:    "Unique identifier for this service",
	}, &serviceName, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	serviceName = strings.TrimSpace(serviceName)

	// Check for conflicts
	if composeFile.ServiceExists(serviceName) {
		if err := handleServiceConflict(composeFile, serviceName); err != nil {
			return err
		}
	}

	service.Name = serviceName

	// Step 2: Image or Build?
	var useImage bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Use a pre-built image? (No = build from Dockerfile)",
		Default: true,
	}, &useImage); err != nil {
		return err
	}

	if useImage {
		// Get image name
		var image string
		if err := survey.AskOne(&survey.Input{
			Message: "Docker image:",
			Default: "nginx:latest",
			Help:    "e.g., nginx:latest, postgres:15, node:20-alpine",
		}, &image, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		service.Image = strings.TrimSpace(image)
	} else {
		// Build configuration
		service.Build = &core.BuildConfig{}

		var buildContext string
		if err := survey.AskOne(&survey.Input{
			Message: "Build context path:",
			Default: ".",
			Help:    "Path to directory containing Dockerfile",
		}, &buildContext); err != nil {
			return err
		}
		service.Build.Context = strings.TrimSpace(buildContext)

		var dockerfile string
		if err := survey.AskOne(&survey.Input{
			Message: "Dockerfile name:",
			Default: "Dockerfile",
		}, &dockerfile); err != nil {
			return err
		}
		service.Build.Dockerfile = strings.TrimSpace(dockerfile)
	}

	// Step 3: Ports
	var addPorts bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Expose ports?",
		Default: true,
	}, &addPorts); err != nil {
		return err
	}

	if addPorts {
		service.Ports = askForPorts()
	}

	// Step 4: Environment Variables
	var addEnv bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Add environment variables?",
		Default: false,
	}, &addEnv); err != nil {
		return err
	}

	if addEnv {
		service.Environment = askForEnvironmentVars()
	}

	// Step 5: Volumes
	var addVolumes bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Mount volumes?",
		Default: false,
	}, &addVolumes); err != nil {
		return err
	}

	if addVolumes {
		service.Volumes = askForVolumeMounts()
	}

	// Step 6: Networks
	var addNetworks bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Connect to networks?",
		Default: false,
	}, &addNetworks); err != nil {
		return err
	}

	if addNetworks {
		service.Networks = askForNetworks(composeFile)
	}

	// Step 7: Dependencies
	var addDependencies bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Add service dependencies (depends_on)?",
		Default: false,
	}, &addDependencies); err != nil {
		return err
	}

	if addDependencies {
		service.DependsOn = askForDependencies(composeFile, serviceName)
	}

	// Step 8: Restart Policy
	var restartPolicy string
	if err := survey.AskOne(&survey.Select{
		Message: "Restart policy:",
		Options: []string{"no", "always", "on-failure", "unless-stopped"},
		Default: "unless-stopped",
		Help:    "Restart policy for the service",
	}, &restartPolicy); err != nil {
		return err
	}
	service.Restart = restartPolicy

	// Step 9: Advanced options
	var configureAdvanced bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure advanced options? (command, working_dir, hostname, etc.)",
		Default: false,
	}, &configureAdvanced); err != nil {
		return err
	}

	if configureAdvanced {
		configureAdvancedOptions(&service)
	}

	// Preview and Confirm
	if err := previewAndConfirmService(composeFile, service, composePath); err != nil {
		return err
	}

	return nil
}

// Helper functions

func askForPorts() []string {
	var ports []string
	fmt.Println("\nEnter port mappings (press Enter with empty value to finish):")
	for {
		var port string
		if err := survey.AskOne(&survey.Input{
			Message: "Port mapping (host:container or container):",
			Help:    "e.g., 8080:80 or 3000",
		}, &port); err != nil || port == "" {
			break
		}
		port = strings.TrimSpace(port)
		if port != "" {
			ports = append(ports, port)
		}

		var addMore bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Add another port?",
			Default: false,
		}, &addMore); err != nil || !addMore {
			break
		}
	}
	return ports
}

func askForEnvironmentVars() map[string]string {
	env := make(map[string]string)
	fmt.Println("\nEnter environment variables (press Enter with empty name to finish):")
	for {
		var key, value string

		if err := survey.AskOne(&survey.Input{
			Message: "Environment variable name:",
			Help:    "e.g., DATABASE_URL, API_KEY",
		}, &key); err != nil || key == "" {
			break
		}

		key = strings.TrimSpace(key)
		if key == "" {
			break
		}

		if err := survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("Value for %s:", key),
		}, &value); err != nil {
			break
		}

		env[key] = value

		var addMore bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Add another variable?",
			Default: false,
		}, &addMore); err != nil || !addMore {
			break
		}
	}
	return env
}

func askForVolumeMounts() []string {
	var volumes []string
	fmt.Println("\nEnter volume mounts (press Enter with empty value to finish):")
	for {
		var volume string
		if err := survey.AskOne(&survey.Input{
			Message: "Volume mount (host:container or volume:container):",
			Help:    "e.g., ./app:/app or data:/var/lib/data",
		}, &volume); err != nil || volume == "" {
			break
		}
		volume = strings.TrimSpace(volume)
		if volume != "" {
			volumes = append(volumes, volume)
		}

		var addMore bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Add another volume?",
			Default: false,
		}, &addMore); err != nil || !addMore {
			break
		}
	}
	return volumes
}

func askForNetworks(composeFile *core.ComposeFile) []string {
	// Get existing networks
	var existingNetworks []string
	for name := range composeFile.Networks {
		existingNetworks = append(existingNetworks, name)
	}

	var networks []string

	if len(existingNetworks) == 0 {
		fmt.Println("\nNo existing networks found. You can create networks later or enter custom network names.")
	}

	fmt.Println("\nEnter network names (press Enter with empty value to finish):")
	for {
		var network string
		if err := survey.AskOne(&survey.Input{
			Message: "Network name:",
			Help:    "Network to connect this service to",
		}, &network); err != nil || network == "" {
			break
		}
		network = strings.TrimSpace(network)
		if network != "" {
			networks = append(networks, network)
		}

		var addMore bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Add another network?",
			Default: false,
		}, &addMore); err != nil || !addMore {
			break
		}
	}
	return networks
}

func askForDependencies(composeFile *core.ComposeFile, currentService string) []string {
	// Get existing services (excluding current one)
	var existingServices []string
	for name := range composeFile.Services {
		if name != currentService {
			existingServices = append(existingServices, name)
		}
	}

	if len(existingServices) == 0 {
		fmt.Println("\nNo other services found to depend on.")
		return nil
	}

	var dependencies []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select services this service depends on:",
		Options: existingServices,
		Help:    "These services will be started before this one",
	}, &dependencies); err != nil {
		return nil
	}

	return dependencies
}

func configureAdvancedOptions(service *core.Service) {
	fmt.Println("\n⚙️  Advanced Configuration\n")

	// Command
	var setCommand bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Override default command?",
		Default: false,
	}, &setCommand); err == nil && setCommand {
		var command string
		if err := survey.AskOne(&survey.Input{
			Message: "Command:",
			Help:    "Command to run when container starts",
		}, &command); err == nil && command != "" {
			service.Command = command
		}
	}

	// Working directory
	var setWorkdir bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Set working directory?",
		Default: false,
	}, &setWorkdir); err == nil && setWorkdir {
		var workdir string
		if err := survey.AskOne(&survey.Input{
			Message: "Working directory:",
			Default: "/app",
		}, &workdir); err == nil && workdir != "" {
			service.WorkingDir = strings.TrimSpace(workdir)
		}
	}

	// User
	var setUser bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Set user?",
		Default: false,
	}, &setUser); err == nil && setUser {
		var user string
		if err := survey.AskOne(&survey.Input{
			Message: "User (uid:gid or username):",
			Help:    "e.g., 1000:1000 or node",
		}, &user); err == nil && user != "" {
			service.User = strings.TrimSpace(user)
		}
	}

	// Hostname
	var setHostname bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Set custom hostname?",
		Default: false,
	}, &setHostname); err == nil && setHostname {
		var hostname string
		if err := survey.AskOne(&survey.Input{
			Message: "Hostname:",
		}, &hostname); err == nil && hostname != "" {
			service.Hostname = strings.TrimSpace(hostname)
		}
	}
}

func previewAndConfirmService(composeFile *core.ComposeFile, service core.Service, composePath string) error {
	// Add service to compose file (temporary for preview)
	composeFile.AddService(service)

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

	fmt.Println("\n✅ Service added successfully!")
	fmt.Printf("   Service '%s' has been added to docker-compose.yml\n", service.Name)
	return nil
}

func handleServiceConflict(composeFile *core.ComposeFile, serviceName string) error {
	var overwrite bool
	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("⚠️  Service '%s' already exists. Overwrite?", serviceName),
		Default: false,
	}, &overwrite); err != nil {
		return err
	}

	if !overwrite {
		return fmt.Errorf("service already exists and overwrite was declined")
	}

	return nil
}
