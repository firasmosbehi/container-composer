package plugins

// Plugin represents a Container Composer plugin
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Initialize initializes the plugin
	Initialize() error

	// Execute executes the plugin
	Execute(args []string) error
}

// Manager manages plugin lifecycle
type Manager struct {
	plugins map[string]Plugin
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a new plugin
func (m *Manager) Register(plugin Plugin) error {
	// TODO: Implement plugin registration
	m.plugins[plugin.Name()] = plugin
	return nil
}

// Load loads plugins from the plugins directory
func (m *Manager) Load(pluginDir string) error {
	// TODO: Implement plugin loading
	return nil
}

// Execute executes a plugin by name
func (m *Manager) Execute(name string, args []string) error {
	// TODO: Implement plugin execution
	plugin, exists := m.plugins[name]
	if !exists {
		return nil
	}
	return plugin.Execute(args)
}

// List returns all registered plugins
func (m *Manager) List() []Plugin {
	plugins := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}