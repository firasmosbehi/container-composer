package templates

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed lamp.yaml
var lampTemplate string

//go:embed lemp.yaml
var lempTemplate string

//go:embed mean.yaml
var meanTemplate string

//go:embed django.yaml
var djangoTemplate string

//go:embed rails.yaml
var railsTemplate string

//go:embed nodejs.yaml
var nodejsTemplate string

//go:embed microservices.yaml
var microservicesTemplate string

// Template represents a project template
type Template struct {
	Name        string
	Description string
	Category    string
	Content     string
	Files       []TemplateFile
}

// TemplateFile represents a file to be created
type TemplateFile struct {
	Path    string
	Content string
}

// Available template categories
const (
	CategoryWeb          = "web"
	CategoryDatabase     = "database"
	CategoryMicroservice = "microservice"
	CategoryFullStack    = "fullstack"
)

// GetAvailableTemplates returns all available project templates
func GetAvailableTemplates() []Template {
	return []Template{
		{
			Name:        "lamp",
			Description: "Linux, Apache, MySQL, PHP - Classic web stack",
			Category:    CategoryFullStack,
			Content:     lampTemplate,
		},
		{
			Name:        "lemp",
			Description: "Linux, Nginx, MySQL, PHP - Modern web stack",
			Category:    CategoryFullStack,
			Content:     lempTemplate,
		},
		{
			Name:        "mean",
			Description: "MongoDB, Express, Angular, Node.js - JavaScript full-stack",
			Category:    CategoryFullStack,
			Content:     meanTemplate,
		},
		{
			Name:        "nodejs",
			Description: "Node.js with PostgreSQL and Redis - Modern backend",
			Category:    CategoryWeb,
			Content:     nodejsTemplate,
		},
		{
			Name:        "django",
			Description: "Django with PostgreSQL and Redis - Python web framework",
			Category:    CategoryWeb,
			Content:     djangoTemplate,
		},
		{
			Name:        "rails",
			Description: "Ruby on Rails with PostgreSQL and Sidekiq",
			Category:    CategoryWeb,
			Content:     railsTemplate,
		},
		{
			Name:        "microservices",
			Description: "Microservices with API Gateway, monitoring, and message queue",
			Category:    CategoryMicroservice,
			Content:     microservicesTemplate,
		},
	}
}

// GetTemplate returns a specific template by name
func GetTemplate(name string) (*Template, error) {
	templates := GetAvailableTemplates()
	for _, tmpl := range templates {
		if tmpl.Name == name {
			return &tmpl, nil
		}
	}
	return nil, fmt.Errorf("template '%s' not found", name)
}

// TemplateVars holds variables for template generation
type TemplateVars struct {
	ProjectName string
}

// Generate generates project files from a template
func (t *Template) Generate(outputDir string, vars TemplateVars) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parse and execute the docker-compose template
	tmpl, err := template.New("compose").Parse(t.Content)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create docker-compose.yml file
	composePath := filepath.Join(outputDir, "docker-compose.yml")
	composeFile, err := os.Create(composePath)
	if err != nil {
		return fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}
	defer composeFile.Close()

	if err := tmpl.Execute(composeFile, vars); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Create .env file
	if err := t.createEnvFile(outputDir); err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}

	// Create README
	if err := t.createReadme(outputDir, vars); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	// Create .gitignore
	if err := t.createGitignore(outputDir); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	// Create necessary directories based on template
	if err := t.createDirectories(outputDir); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	return nil
}

// createEnvFile creates a .env file with default values
func (t *Template) createEnvFile(outputDir string) error {
	envPath := filepath.Join(outputDir, ".env.example")

	var envContent string
	switch t.Name {
	case "lamp", "lemp":
		envContent = `# Database Configuration
DB_ROOT_PASSWORD=rootpassword
DB_DATABASE=myapp
DB_USER=dbuser
DB_PASSWORD=dbpassword
`
	case "mean":
		envContent = `# MongoDB Configuration
MONGO_USER=admin
MONGO_PASSWORD=password
`
	case "django", "rails", "nodejs":
		envContent = `# Database Configuration
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=postgres

# Application Configuration
DEBUG=True
`
	case "microservices":
		envContent = `# Service Configuration
POSTGRES_PASSWORD=postgres
GRAFANA_PASSWORD=admin
`
	default:
		envContent = "# Environment variables\n"
	}

	return os.WriteFile(envPath, []byte(envContent), 0644)
}

// createReadme creates a README file with instructions
func (t *Template) createReadme(outputDir string, vars TemplateVars) error {
	readmePath := filepath.Join(outputDir, "README.md")

	readmeContent := fmt.Sprintf("# %s\n\n"+
		"This project was generated using Container Composer with the **%s** template.\n\n"+
		"## Description\n\n"+
		"%s\n\n"+
		"## Getting Started\n\n"+
		"### Prerequisites\n\n"+
		"- Docker\n"+
		"- Docker Compose\n"+
		"- Container Composer (optional, for enhanced features)\n\n"+
		"### Quick Start\n\n"+
		"1. Copy the example environment file:\n"+
		"   ```bash\n"+
		"   cp .env.example .env\n"+
		"   ```\n\n"+
		"2. Edit the `.env` file with your configuration\n\n"+
		"3. Start the services:\n"+
		"   ```bash\n"+
		"   docker-compose up -d\n"+
		"   # or\n"+
		"   container-composer up\n"+
		"   ```\n\n"+
		"4. View logs:\n"+
		"   ```bash\n"+
		"   docker-compose logs -f\n"+
		"   # or\n"+
		"   container-composer logs --follow\n"+
		"   ```\n\n"+
		"5. Stop the services:\n"+
		"   ```bash\n"+
		"   docker-compose down\n"+
		"   # or\n"+
		"   container-composer down\n"+
		"   ```\n\n"+
		"## Services\n\n"+
		"Check your `docker-compose.yml` file to see all configured services and their ports.\n\n"+
		"## Container Composer Features\n\n"+
		"If you have Container Composer installed, you can use these enhanced features:\n\n"+
		"```bash\n"+
		"# View service status and health\n"+
		"container-composer status\n\n"+
		"# Access a service shell\n"+
		"container-composer shell <service-name>\n\n"+
		"# View filtered logs\n"+
		"container-composer logs --filter \"error|warning\"\n"+
		"```\n\n"+
		"## License\n\n"+
		"MIT\n",
		vars.ProjectName, t.Name, t.Description)

	return os.WriteFile(readmePath, []byte(readmeContent), 0644)
}

// createGitignore creates a .gitignore file
func (t *Template) createGitignore(outputDir string) error {
	gitignorePath := filepath.Join(outputDir, ".gitignore")

	gitignoreContent := `# Environment files
.env
.env.local

# Docker
docker-compose.override.yml

# Logs
*.log

# OS files
.DS_Store
Thumbs.db

# IDE
.vscode/
.idea/
*.swp
*.swo
`

	return os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
}

// createDirectories creates necessary directories for the template
func (t *Template) createDirectories(outputDir string) error {
	var dirs []string

	switch t.Name {
	case "lamp":
		dirs = []string{"src"}
	case "lemp":
		dirs = []string{"src", "nginx/conf.d"}
	case "mean":
		dirs = []string{"frontend", "backend"}
	case "django":
		dirs = []string{"app"}
	case "rails":
		dirs = []string{"app"}
	case "nodejs":
		dirs = []string{"app"}
	case "microservices":
		dirs = []string{
			"services/user-service",
			"services/product-service",
			"services/order-service",
			"gateway",
			"monitoring",
			"postgres/init",
		}
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(outputDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}

		// Create a placeholder .gitkeep file
		gitkeepPath := filepath.Join(dirPath, ".gitkeep")
		if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil {
			return err
		}
	}

	return nil
}
