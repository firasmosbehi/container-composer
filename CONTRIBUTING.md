# Contributing to Container Composer

Thank you for your interest in contributing to Container Composer! This document provides guidelines and instructions for contributing.

## Development Setup

1. Clone the repository:
```bash
git clone https://github.com/firasmosbahi/container-composer.git
cd container-composer
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
make build
```

4. Run tests:
```bash
make test
```

## Project Structure

```
container-composer/
├── cmd/           # Application entry points
├── cli/           # CLI commands and interface
├── core/          # Core Docker/Compose functionality
├── tui/           # Terminal UI components
├── plugins/       # Plugin system
├── templates/     # Project templates
├── utils/         # Utility functions
└── internal/      # Internal packages (config, etc.)
```

## Development Workflow

1. Create a new branch for your feature/fix:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and ensure code quality:
```bash
make lint
make test
```

3. Commit your changes with clear messages:
```bash
git commit -m "Add feature: description"
```

4. Push to your fork and create a Pull Request

## Code Style

- Follow standard Go conventions and idioms
- Run `go fmt` before committing
- Add comments for exported functions and types
- Keep functions focused and modular
- Write tests for new functionality

## Testing

- Write unit tests for new features
- Ensure all tests pass before submitting PR
- Aim for high test coverage on core functionality
- Include integration tests for complex features

## Pull Request Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Ensure CI passes
- Update documentation if needed
- Keep PRs focused on a single feature/fix

## Feature Requests and Bug Reports

- Use GitHub Issues for bug reports and feature requests
- Provide clear reproduction steps for bugs
- Include system information (OS, Docker version, etc.)
- Check existing issues before creating new ones

## Code Review Process

- All submissions require review
- Maintainers will provide feedback
- Address review comments promptly
- Be open to suggestions and improvements

## License

By contributing, you agree that your contributions will be licensed under the MIT License.