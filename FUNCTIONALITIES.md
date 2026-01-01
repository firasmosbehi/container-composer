# Container Composer - User Guide

This document describes the fully implemented functionalities of Container Composer and how to use them.

## Table of Contents

1. [Overview](#overview)
2. [Installation](#installation)
3. [Available Commands](#available-commands)
4. [Project Initialization](#project-initialization)
5. [Available Templates](#available-templates)
6. [Usage Examples](#usage-examples)

---

## Overview

Container Composer is a CLI tool for creating and managing Docker Compose projects. It provides production-ready templates for common technology stacks, making it easy to bootstrap new projects with best practices built-in.

**Current Capabilities:**

- Initialize new Docker Compose projects from templates
- Choose from 7 production-ready stack templates
- Auto-generate project structure with all necessary files
- Interactive and non-interactive modes

---

## Installation

### Building from Source

```bash
# Navigate to the project directory
cd /path/to/container-composer

# Build the binary
make build

# The executable will be at ./bin/container-composer
./bin/container-composer --version

# Optional: Install to system
make install
```

### Cross-Platform Builds

The project supports building for multiple platforms:

- Linux (amd64, arm64)
- macOS/Darwin (amd64, arm64)
- Windows (amd64)

---

## Available Commands

### Global Flags

These flags work with all commands:

| Flag | Description |
| --- | --- |
| `--verbose` | Enable verbose output for detailed logging |
| `--debug` | Enable debug mode with extra diagnostic information |
| `--help`, `-h` | Display help information for any command |
| `--version`, `-v` | Show version, build date, and commit information |

### Commands

| Command | Description | Status |
| --- | --- | --- |
| `init` | Initialize a new Docker Compose project | ✅ Implemented |
| `help` | Display help information | ✅ Implemented |

---

## Project Initialization

The `init` command creates a new Docker Compose project from a template.

### Command Syntax

```bash
container-composer init [project-name] [flags]
```

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--template`, `-t` | Specify template to use (lamp, lemp, mean, nodejs, django, rails, microservices) | Interactive prompt |
| `--no-prompt` | Skip interactive prompts (requires --template) | false |

### What Gets Generated

When you initialize a project, Container Composer creates:

1. **docker-compose.yml** - Complete service definitions with:
   - Service configurations
   - Network setup
   - Volume definitions
   - Health checks (where applicable)
   - Environment variables

2. **.env.example** - Template for environment variables with:
   - Database credentials
   - Port configurations
   - Service-specific settings

3. **README.md** - Project documentation with:
   - Quick start guide
   - Service descriptions
   - Port mappings
   - Common commands

4. **.gitignore** - Git exclusions for:
   - Environment files
   - Log files
   - Data directories

5. **Directory Structure** - Project-specific folders based on the selected template

### Interactive Mode

```bash
# Start interactive wizard
container-composer init

# Or specify project name
container-composer init my-project
```

The wizard will:
1. Prompt for project name (if not provided)
2. Display available templates
3. Let you select a template
4. Generate all project files

### Non-Interactive Mode

```bash
# Create project with specific template
container-composer init my-project --template=nodejs --no-prompt

# Using short flag
container-composer init my-api -t microservices --no-prompt
```

---

## Available Templates

### 1. LAMP Stack

**Use Case:** Traditional PHP applications, WordPress, content management systems

**Services:**
- Apache 2.4 (web server with mod_rewrite)
- MySQL 8.0 (database with persistent storage)
- PHP 8.1 (with common extensions)
- phpMyAdmin (database management UI)

**Template ID:** `lamp`

**Default Ports:**
- Apache: 8080
- MySQL: 3306
- phpMyAdmin: 8081

**Directory Structure:**
```
src/          # PHP application code
logs/         # Apache logs
mysql-data/   # Database persistent storage
```

---

### 2. LEMP Stack

**Use Case:** High-performance PHP applications, Laravel, Symfony

**Services:**
- Nginx (optimized web server)
- MySQL 8.0 (database)
- PHP-FPM (FastCGI Process Manager)
- phpMyAdmin (database management)

**Template ID:** `lemp`

**Default Ports:**
- Nginx: 8080
- MySQL: 3306
- phpMyAdmin: 8081

**Directory Structure:**
```
src/          # PHP application code
nginx/        # Nginx configuration
mysql-data/   # Database persistent storage
```

---

### 3. MEAN Stack

**Use Case:** Modern JavaScript applications, Single Page Applications (SPAs)

**Services:**
- MongoDB (NoSQL database with authentication)
- Express (Node.js web framework)
- Angular (frontend framework with dev server)
- Node.js (backend runtime)

**Template ID:** `mean`

**Default Ports:**
- Frontend (Angular): 4200
- Backend (Node.js): 3000
- MongoDB: 27017

**Directory Structure:**
```
frontend/     # Angular application
backend/      # Node.js/Express API
mongo-data/   # MongoDB persistent storage
```

---

### 4. Node.js Stack

**Use Case:** RESTful APIs, real-time applications, microservices

**Services:**
- Node.js 18 LTS (application runtime)
- PostgreSQL 15 (relational database)
- Redis (caching and session storage)

**Template ID:** `nodejs`

**Default Ports:**
- Node.js: 3000
- PostgreSQL: 5432
- Redis: 6379

**Directory Structure:**
```
src/          # Node.js application code
postgres-data/ # PostgreSQL data
redis-data/   # Redis data
```

---

### 5. Django Stack

**Use Case:** Python web applications, data-driven websites, REST APIs

**Services:**
- Django (Python web framework with auto-reload)
- PostgreSQL (primary database)
- Redis (caching)
- Celery (background task processing)
- Celery Beat (scheduled tasks)

**Template ID:** `django`

**Default Ports:**
- Django: 8000
- PostgreSQL: 5432
- Redis: 6379

**Directory Structure:**
```
app/          # Django application
postgres-data/ # Database storage
redis-data/   # Redis storage
```

---

### 6. Rails Stack

**Use Case:** Ruby web applications, rapid development

**Services:**
- Ruby on Rails (with hot-reload)
- PostgreSQL (database)
- Redis (caching and job queue)
- Sidekiq (background job processing)

**Template ID:** `rails`

**Default Ports:**
- Rails: 3000
- PostgreSQL: 5432
- Redis: 6379

**Directory Structure:**
```
app/          # Rails application
postgres-data/ # Database storage
redis-data/   # Redis storage
```

---

### 7. Microservices

**Use Case:** Distributed systems, scalable architectures, service-oriented applications

**Services:**
- API Gateway (routing and load balancing)
- User Service (user management microservice)
- Product Service (product catalog microservice)
- Order Service (order processing microservice)
- RabbitMQ (message queue)
- PostgreSQL (shared database)
- Redis (distributed cache)
- Prometheus (metrics collection)
- Grafana (monitoring dashboard)

**Template ID:** `microservices`

**Default Ports:**
- API Gateway: 8080
- User Service: 3001
- Product Service: 3002
- Order Service: 3003
- RabbitMQ: 5672 (AMQP), 15672 (Management UI)
- PostgreSQL: 5432
- Redis: 6379
- Prometheus: 9090
- Grafana: 3000

**Directory Structure:**
```
gateway/      # API Gateway
services/
  user/       # User microservice
  product/    # Product microservice
  order/      # Order microservice
postgres-data/ # Database storage
rabbitmq-data/ # Message queue data
prometheus/   # Prometheus configuration
grafana/      # Grafana configuration
```

---

## Usage Examples

### Example 1: Create a Node.js API Project

```bash
# Interactive mode
container-composer init my-api
# Select "nodejs" from the template list

# Non-interactive mode
container-composer init my-api --template=nodejs --no-prompt

# Navigate to project
cd my-api

# Review the generated files
ls -la

# Start the services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

### Example 2: Create a Django Web Application

```bash
# Create project
container-composer init my-webapp -t django --no-prompt

# Navigate and start
cd my-webapp
docker-compose up -d

# Access Django at http://localhost:8000
# PostgreSQL available at localhost:5432
# Redis available at localhost:6379
```

### Example 3: Create a Microservices Architecture

```bash
# Create project
container-composer init my-platform -t microservices --no-prompt

# Navigate to project
cd my-platform

# Review the architecture
cat README.md

# Start all services
docker-compose up -d

# Access API Gateway at http://localhost:8080
# View Grafana dashboard at http://localhost:3000
# RabbitMQ management at http://localhost:15672
```

### Example 4: Create a WordPress Site (LAMP)

```bash
# Create project
container-composer init my-blog --template=lamp --no-prompt

# Navigate and start
cd my-blog
docker-compose up -d

# Access Apache at http://localhost:8080
# phpMyAdmin at http://localhost:8081
```

### Example 5: Using Verbose Mode

```bash
# See detailed output during project creation
container-composer --verbose init my-project -t nodejs --no-prompt

# Or use debug mode for maximum detail
container-composer --debug init my-project -t rails --no-prompt
```

---

## Next Steps After Initialization

After creating a project with `container-composer init`, follow these steps:

### 1. Review Generated Files

```bash
cd your-project-name
ls -la
cat README.md
```

### 2. Configure Environment Variables

```bash
# Copy the example environment file
cp .env.example .env

# Edit with your specific values
nano .env  # or vim, code, etc.
```

### 3. Start the Services

```bash
# Start all services in detached mode
docker-compose up -d

# Or start with logs visible
docker-compose up
```

### 4. Verify Services Are Running

```bash
# Check service status
docker-compose ps

# View logs
docker-compose logs

# Follow logs in real-time
docker-compose logs -f
```

### 5. Access Your Application

Check the README.md in your project for specific URLs and ports for each service.

### 6. Stop Services When Done

```bash
# Stop services
docker-compose down

# Stop and remove volumes (careful: deletes data)
docker-compose down -v
```

---

## Getting Help

### Display General Help

```bash
container-composer --help
```

### Display Help for Specific Command

```bash
container-composer init --help
```

### Check Version

```bash
container-composer --version
```

Output includes:
- Version number
- Git commit hash
- Build date
- Go version used

---

## Best Practices

### 1. Environment Variables

- Never commit `.env` files to version control
- Always use `.env.example` as a template
- Store sensitive data (passwords, API keys) in `.env`

### 2. Data Persistence

- Generated `docker-compose.yml` files include named volumes for data persistence
- Use `docker-compose down` to stop services without losing data
- Use `docker-compose down -v` only when you want to delete all data

### 3. Port Conflicts

- Default ports are configured in the templates
- If ports are already in use, edit `docker-compose.yml` or `.env` to change them
- Check the generated README.md for port mappings

### 4. Customization

- Templates are starting points - customize them for your needs
- Add or remove services in `docker-compose.yml`
- Adjust resource limits, networks, and volumes as needed

---

## Troubleshooting

### Template Not Found

```bash
# Error: unknown template 'xyz'
# Solution: Use one of the valid templates:
container-composer init myapp -t nodejs  # Valid
container-composer init myapp -t lamp    # Valid
container-composer init myapp -t xyz     # Invalid
```

Valid templates: `lamp`, `lemp`, `mean`, `nodejs`, `django`, `rails`, `microservices`

### Directory Already Exists

```bash
# Error: directory already exists
# Solution: Choose a different name or remove the existing directory
rm -rf existing-project
container-composer init existing-project -t nodejs
```

### Missing Template Flag in Non-Interactive Mode

```bash
# Error: --no-prompt requires --template
# Solution: Provide both flags
container-composer init myapp --template=nodejs --no-prompt
```

---

## Technical Details

### Template System

- Templates are embedded in the binary using Go's `//go:embed` directive
- No external files needed at runtime
- Templates use Go's `text/template` for variable substitution
- Variables like `{{.ProjectName}}` are replaced during generation

### Supported Platforms

Binaries are available for:
- **Linux:** amd64, arm64
- **macOS/Darwin:** amd64 (Intel), arm64 (Apple Silicon)
- **Windows:** amd64

### Build Information

Version information is embedded at build time using Go's linker flags:
- Version number from git tags
- Commit hash from git
- Build timestamp
- Go compiler version

---

*Last Updated: 2026-01-01*

*For issues, feature requests, or contributions, visit the project repository.*