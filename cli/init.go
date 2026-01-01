package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/firasmosbahi/container-composer/templates"
	"github.com/spf13/cobra"
)

var (
	initTemplate string
	initNoPrompt bool
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Container Composer project",
	Long: `Initialize a new Container Composer project with an interactive wizard.
You can choose from various templates like LAMP, MEAN, microservices, etc.

Examples:
  container-composer init                     # Interactive mode
  container-composer init my-project          # Interactive mode with project name
  container-composer init --template=lamp     # Use LAMP template directly`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "", "template to use (skip wizard)")
	initCmd.Flags().BoolVar(&initNoPrompt, "no-prompt", false, "skip all prompts and use defaults")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine project name and directory
	projectName := "my-project"
	projectDir := "."

	if len(args) > 0 {
		projectName = args[0]
		projectDir = args[0]
	}

	// If no template specified and not in no-prompt mode, run interactive wizard
	if initTemplate == "" && !initNoPrompt {
		selectedTemplate, selectedProjectName, err := runWizard(projectName)
		if err != nil {
			return err
		}
		initTemplate = selectedTemplate
		if selectedProjectName != "" {
			projectName = selectedProjectName
			if len(args) == 0 {
				projectDir = selectedProjectName
			}
		}
	}

	// If still no template, default to nodejs
	if initTemplate == "" {
		initTemplate = "nodejs"
	}

	// Get the template
	tmpl, err := templates.GetTemplate(initTemplate)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Check if directory exists and is not empty
	if projectDir != "." {
		if _, err := os.Stat(projectDir); err == nil {
			// Directory exists, check if empty
			entries, err := os.ReadDir(projectDir)
			if err != nil {
				return fmt.Errorf("failed to read directory: %w", err)
			}
			if len(entries) > 0 && !initNoPrompt {
				overwrite := false
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("Directory '%s' is not empty. Continue anyway?", projectDir),
					Default: false,
				}
				if err := survey.AskOne(prompt, &overwrite); err != nil {
					return err
				}
				if !overwrite {
					return fmt.Errorf("initialization cancelled")
				}
			}
		}
	}

	// Generate project from template
	fmt.Printf("\n🚀 Initializing project '%s' with template '%s'...\n\n", projectName, initTemplate)

	vars := templates.TemplateVars{
		ProjectName: projectName,
	}

	if err := tmpl.Generate(projectDir, vars); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Print success message
	printSuccessMessage(projectName, projectDir, tmpl)

	return nil
}

func runWizard(defaultProjectName string) (string, string, error) {
	fmt.Println("\n✨ Welcome to Container Composer Project Wizard! ✨\n")

	// Get available templates
	availableTemplates := templates.GetAvailableTemplates()

	// Create template options grouped by category
	templateOptions := make([]string, len(availableTemplates))
	templateMap := make(map[string]*templates.Template)

	for i, tmpl := range availableTemplates {
		option := fmt.Sprintf("%s - %s", tmpl.Name, tmpl.Description)
		templateOptions[i] = option
		templateMap[option] = &availableTemplates[i]
	}

	// Questions
	var selectedOption string
	var projectName string

	questions := []*survey.Question{
		{
			Name: "template",
			Prompt: &survey.Select{
				Message: "Choose a project template:",
				Options: templateOptions,
				Help:    "Select the stack that best fits your project needs",
			},
		},
		{
			Name: "projectname",
			Prompt: &survey.Input{
				Message: "Project name:",
				Default: defaultProjectName,
				Help:    "This will be used in container names and the README",
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Template    string
		ProjectName string
	}{}

	err := survey.Ask(questions, &answers)
	if err != nil {
		return "", "", err
	}

	selectedOption = answers.Template
	projectName = strings.TrimSpace(answers.ProjectName)

	// Extract template name from option
	selectedTemplate := templateMap[selectedOption]
	if selectedTemplate == nil {
		return "", "", fmt.Errorf("invalid template selection")
	}

	// Confirmation
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Create project '%s' with template '%s'?", projectName, selectedTemplate.Name),
		Default: true,
	}

	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return "", "", err
	}

	if !confirm {
		return "", "", fmt.Errorf("initialization cancelled by user")
	}

	return selectedTemplate.Name, projectName, nil
}

func printSuccessMessage(projectName, projectDir string, tmpl *templates.Template) {
	fmt.Println("✅ Project initialized successfully!")
	fmt.Println("\n📁 Created files:")
	fmt.Println("   - docker-compose.yml")
	fmt.Println("   - .env.example")
	fmt.Println("   - README.md")
	fmt.Println("   - .gitignore")

	fmt.Println("\n🎯 Next steps:")

	if projectDir != "." {
		fmt.Printf("   1. cd %s\n", projectDir)
		fmt.Println("   2. cp .env.example .env")
		fmt.Println("   3. Edit .env with your configuration")
		fmt.Println("   4. container-composer up")
	} else {
		fmt.Println("   1. cp .env.example .env")
		fmt.Println("   2. Edit .env with your configuration")
		fmt.Println("   3. container-composer up")
	}

	fmt.Println("\n📚 Template info:")
	fmt.Printf("   Name: %s\n", tmpl.Name)
	fmt.Printf("   Description: %s\n", tmpl.Description)
	fmt.Printf("   Category: %s\n", tmpl.Category)

	fmt.Println("\n💡 Tip: Run 'container-composer status' to check service health after starting!")
	fmt.Println()
}
