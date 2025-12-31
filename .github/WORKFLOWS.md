# GitHub Actions Workflows

This document explains the GitHub Actions workflows configured for Container Composer.

## Workflows Overview

### 1. Release Workflow (`release.yml`)

**Trigger:** Pushes to `main` branch or version tags (`v*`)

**Purpose:** Automatically builds and releases Container Composer binaries for multiple platforms.

**Process:**
1. **Create Release Job**
   - Generates version number from git tags or commit SHA
   - Creates a changelog from commit history
   - Creates a new GitHub release

2. **Build and Upload Job**
   - Builds binaries for multiple platforms:
     - Linux (amd64, arm64)
     - macOS (amd64, arm64)
     - Windows (amd64)
   - Creates compressed archives (.tar.gz for Unix, .zip for Windows)
   - Generates SHA256 checksums for verification
   - Uploads all artifacts to the GitHub release

**Artifacts Produced:**
- `container-composer-linux-amd64.tar.gz` + checksum
- `container-composer-linux-arm64.tar.gz` + checksum
- `container-composer-darwin-amd64.tar.gz` + checksum
- `container-composer-darwin-arm64.tar.gz` + checksum
- `container-composer-windows-amd64.exe.zip` + checksum

### 2. CI Workflow (`ci.yml`)

**Trigger:** Pull requests to `main` and pushes to non-main branches

**Purpose:** Continuous integration testing to ensure code quality.

**Jobs:**

1. **Test Job**
   - Runs all Go tests with race detection
   - Generates code coverage reports
   - Uploads coverage to Codecov

2. **Lint Job**
   - Runs golangci-lint for code quality checks
   - Enforces Go best practices and style guidelines

3. **Build Job**
   - Verifies that the code builds successfully for all target platforms
   - Matrix build across Linux, macOS, Windows and amd64, arm64 architectures

## Versioning

The release workflow uses the following versioning strategy:

- **Tagged releases**: If you push a git tag like `v1.0.0`, it uses that version
- **Automated versions**: On regular pushes to main without tags, generates version like `0.0.0-abc12345` (using commit SHA)

To create a version release:
```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

## Dependabot

Configured in `.github/dependabot.yml` to automatically:
- Update Go module dependencies weekly
- Update GitHub Actions versions weekly
- Create pull requests for dependency updates

## Security

All workflows run with minimal permissions:
- `contents: write` only for release workflow (to create releases)
- No other elevated permissions granted

## Usage

### Creating a Release

**Option 1: Tag-based release (Recommended for stable versions)**
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

**Option 2: Push to main (For continuous releases)**
```bash
git push origin main
```

Both methods will trigger the release workflow automatically.

### Monitoring Builds

- Check the [Actions tab](https://github.com/firasmosbahi/container-composer/actions) in GitHub
- View build logs for debugging
- Download artifacts from completed workflow runs

## Customization

### Adding New Platforms

To add support for additional platforms, edit `.github/workflows/release.yml`:

```yaml
matrix:
  include:
    # Add new platform configuration
    - os: ubuntu-latest
      goos: freebsd
      goarch: amd64
      artifact_name: container-composer
      asset_name: container-composer-freebsd-amd64
```

### Changing Go Version

Update the `go-version` in both `release.yml` and `ci.yml`:

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.22'  # Change version here
```

## Troubleshooting

### Release Not Creating

- Ensure you have pushed to the `main` branch
- Check that GitHub Actions is enabled for the repository
- Verify that `GITHUB_TOKEN` has proper permissions

### Build Failing

- Check the Actions tab for detailed error logs
- Run `make test` and `make build` locally to reproduce issues
- Ensure all tests pass before pushing

### Artifacts Not Uploading

- Verify the release was created successfully in the first job
- Check that the artifact paths in the workflow are correct
- Ensure file permissions allow reading the built binaries