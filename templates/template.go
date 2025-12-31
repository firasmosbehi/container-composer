package templates

// Template represents a project template
type Template struct {
	Name        string
	Description string
	Category    string
	Files       map[string]string
}

// Available template categories
const (
	CategoryWeb         = "web"
	CategoryDatabase    = "database"
	CategoryMicroservice = "microservice"
	CategoryFullStack   = "fullstack"
)

// GetAvailableTemplates returns all available project templates
func GetAvailableTemplates() []Template {
	return []Template{
		{
			Name:        "lamp",
			Description: "Linux, Apache, MySQL, PHP stack",
			Category:    CategoryFullStack,
		},
		{
			Name:        "mean",
			Description: "MongoDB, Express, Angular, Node.js stack",
			Category:    CategoryFullStack,
		},
		{
			Name:        "lemp",
			Description: "Linux, Nginx, MySQL, PHP stack",
			Category:    CategoryFullStack,
		},
		{
			Name:        "django",
			Description: "Django with PostgreSQL and Redis",
			Category:    CategoryWeb,
		},
		{
			Name:        "rails",
			Description: "Ruby on Rails with PostgreSQL",
			Category:    CategoryWeb,
		},
		{
			Name:        "microservices",
			Description: "Microservices architecture with API gateway",
			Category:    CategoryMicroservice,
		},
	}
}

// GetTemplate returns a specific template by name
func GetTemplate(name string) (*Template, error) {
	// TODO: Implement template retrieval
	return nil, nil
}

// Generate generates project files from a template
func (t *Template) Generate(outputDir string, vars map[string]string) error {
	// TODO: Implement template generation
	return nil
}