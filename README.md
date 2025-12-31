<div align="center">
  <img src="assets/logo/banner.svg" alt="Container Composer" width="800"/>

  <p align="center">
    <strong>Orchestrate your Docker Compose projects with ease</strong>
  </p>

  <p align="center">
    <a href="https://github.com/firasmosbahi/container-composer/actions/workflows/release.yml">
      <img src="https://github.com/firasmosbahi/container-composer/actions/workflows/release.yml/badge.svg" alt="Release">
    </a>
    <a href="https://github.com/firasmosbahi/container-composer/actions/workflows/ci.yml">
      <img src="https://github.com/firasmosbahi/container-composer/actions/workflows/ci.yml/badge.svg" alt="CI">
    </a>
    <a href="https://goreportcard.com/report/github.com/firasmosbahi/container-composer">
      <img src="https://goreportcard.com/badge/github.com/firasmosbahi/container-composer" alt="Go Report Card">
    </a>
    <a href="https://opensource.org/licenses/MIT">
      <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT">
    </a>
  </p>
</div>

---

A powerful, developer-friendly CLI tool for managing and maintaining Docker Compose projects.

## Overview

Container Composer enhances your Docker Compose workflow with intelligent features, better debugging capabilities, and an improved developer experience. It works seamlessly with your existing `docker-compose.yml` files without modification.

## Features

### Project Management
- Interactive project initialization wizard with templates (LAMP, MEAN, microservices, etc.)
- Dependency graph visualization showing service relationships
- Environment variable management across multiple .env files
- Secrets management integration (Vault, SOPS)
- Multi-environment support (dev, staging, prod) with easy switching

### Development & Debugging
- Live log aggregation from all services with filtering, search, and highlighting
- Service health monitoring dashboard
- Port conflict detection and resolution
- Volume inspection and cleanup tools
- Network debugging (show connections between services)
- Shell access manager (quick exec into any service)
- Hot-reload configuration (detect compose file changes and prompt for reload)

### Quality of Life
- Smart docker-compose command wrapper with better error messages
- Dependency ordering visualization and validation
- Resource usage monitoring per service (CPU, memory, network)
- Quick service restart/rebuild without affecting others
- Backup/restore functionality for volumes
- Import/export configurations

### Advanced Features
- Compose file linting and best practices checker
- Security scanning for images
- Auto-scaling simulation/testing
- Performance profiling tools
- CI/CD pipeline generator
- Migration tools (convert from docker run commands, Kubernetes, etc.)
- Plugin system for extensibility

## Installation

### From Release (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/firasmosbahi/container-composer/releases):

#### Linux / macOS
```bash
# Download the latest release (replace VERSION and OS/ARCH as needed)
curl -LO https://github.com/firasmosbahi/container-composer/releases/latest/download/container-composer-linux-amd64.tar.gz

# Extract the archive
tar xzf container-composer-linux-amd64.tar.gz

# Move to a directory in your PATH
sudo mv container-composer /usr/local/bin/

# Verify installation
container-composer --version
```

#### Windows
Download the `.zip` file from the releases page and extract it to a directory in your PATH.

### From Source

```bash
go install github.com/firasmosbahi/container-composer/cmd/container-composer@latest
```

### Verify Installation

After installation, verify the checksum (optional but recommended):
```bash
# Download the checksum file
curl -LO https://github.com/firasmosbahi/container-composer/releases/latest/download/container-composer-linux-amd64.tar.gz.sha256

# Verify
sha256sum -c container-composer-linux-amd64.tar.gz.sha256
```

## Quick Start

```bash
# Initialize a new project
container-composer init

# Start services
container-composer up

# View logs with filtering
container-composer logs --filter "error|warning"

# Monitor service health
container-composer status

# Access service shell
container-composer shell <service-name>
```

## Architecture

Container Composer is built with a modular architecture:

```
container-composer/
├── core/          # Docker interaction, compose parsing
├── cli/           # Command-line interface
├── tui/           # Terminal UI components
├── plugins/       # Plugin system
├── templates/     # Project templates
└── utils/         # Helpers, validators
```

## Design Principles

- **Offline-first**: Local operations never require internet
- **Non-invasive**: Doesn't modify compose files unless explicitly requested
- **Backward compatible**: Works with standard docker-compose files
- **Extensive logging**: Debug mode for troubleshooting
- **Developer experience**: Make common tasks 10x easier

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] Core project structure and CLI framework
- [ ] Docker Compose file parsing and validation
- [ ] Basic service management (up, down, restart)
- [ ] Log aggregation and filtering
- [ ] Service health monitoring
- [ ] Interactive TUI dashboard
- [ ] Plugin system
- [ ] Project templates
- [ ] Advanced features (security scanning, migration tools, etc.)