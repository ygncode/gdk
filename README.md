# GDK - GitHub Deploy Key Manager

[![CI](https://github.com/ygncode/gdk/actions/workflows/ci.yml/badge.svg)](https://github.com/ygncode/gdk/actions/workflows/ci.yml)
[![Release](https://github.com/ygncode/gdk/actions/workflows/release.yml/badge.svg)](https://github.com/ygncode/gdk/actions/workflows/release.yml)

A CLI tool for managing multiple GitHub Deploy Keys on a single machine. Automates SSH key creation and `~/.ssh/config` management.

## Features

- Generate Ed25519 SSH keys for each repository
- Automatically configure `~/.ssh/config` with unique host aliases
- Support for multiple deploy keys on a single machine
- Simple commands: `add`, `list`, `show`, `remove`, `update`
- Self-update capability

## Installation

### Using wget (Linux/macOS)

```bash
# Linux (amd64)
wget https://github.com/ygncode/gdk/releases/latest/download/gdk-linux-amd64 -O gdk
chmod +x gdk
sudo mv gdk /usr/local/bin/

# Linux (arm64)
wget https://github.com/ygncode/gdk/releases/latest/download/gdk-linux-arm64 -O gdk
chmod +x gdk
sudo mv gdk /usr/local/bin/

# macOS (Intel)
wget https://github.com/ygncode/gdk/releases/latest/download/gdk-darwin-amd64 -O gdk
chmod +x gdk
sudo mv gdk /usr/local/bin/

# macOS (Apple Silicon)
wget https://github.com/ygncode/gdk/releases/latest/download/gdk-darwin-arm64 -O gdk
chmod +x gdk
sudo mv gdk /usr/local/bin/
```

### Using curl

```bash
# Linux (amd64)
curl -L https://github.com/ygncode/gdk/releases/latest/download/gdk-linux-amd64 -o gdk
chmod +x gdk
sudo mv gdk /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/ygncode/gdk/releases/latest/download/gdk-darwin-arm64 -o gdk
chmod +x gdk
sudo mv gdk /usr/local/bin/
```

### Using Go

```bash
go install github.com/ygncode/gdk@latest
```

### From Source

```bash
git clone https://github.com/ygncode/gdk.git
cd gdk
go build -o gdk .
sudo mv gdk /usr/local/bin/
```

## Usage

### Add a deploy key

```bash
gdk add myorg/private-repo
```

Output:
```
Deploy key created successfully!

Public key (add this to GitHub deploy keys):
==================================================
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... gdk:myorg/private-repo
==================================================

Clone command:
  git clone git@github.com-myorg-private-repo:myorg/private-repo.git

For existing repos, update remote:
  git remote set-url origin git@github.com-myorg-private-repo:myorg/private-repo.git
```

Then add the public key to your GitHub repository:
1. Go to your repository on GitHub
2. Navigate to **Settings** > **Deploy keys** > **Add deploy key**
3. Paste the public key and save

### List all deploy keys

```bash
gdk list
```

Output:
```
REPOSITORY           HOST ALIAS                    CREATED
----------           ----------                    -------
myorg/private-repo   github.com-myorg-private-repo 2025-01-15
myorg/another-repo   github.com-myorg-another-repo 2025-01-15
```

JSON output:
```bash
gdk list --format json
```

### Remove a deploy key

```bash
gdk remove myorg/private-repo
```

Skip confirmation:
```bash
gdk remove myorg/private-repo --yes
```

### Overwrite existing key

```bash
gdk add myorg/private-repo --force
```

### Show deploy key details

```bash
gdk show myorg/private-repo
```

Output:
```
Repository:     myorg/private-repo
Host Alias:     github.com-myorg-private-repo
Created:        2025-01-15 10:30:00

SSH Keys:
  Private Key:  /home/user/.ssh/deploy-keys/myorg-private-repo/id_ed25519
  Public Key:   /home/user/.ssh/deploy-keys/myorg-private-repo/id_ed25519.pub

Public Key Content:
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... gdk:myorg/private-repo

Git Commands:
  Clone:        git clone git@github.com-myorg-private-repo:myorg/private-repo.git
  Set Remote:   git remote set-url origin git@github.com-myorg-private-repo:myorg/private-repo.git
```

### Update gdk

Update to the latest version:
```bash
gdk update
```

Check for updates without installing:
```bash
gdk update --check
```

## How It Works

1. **Key Generation**: Creates an Ed25519 SSH key pair for each repository
2. **Key Storage**: Saves keys in `~/.ssh/deploy-keys/<owner>-<repo>/`
3. **SSH Config**: Adds a host alias to `~/.ssh/config`:

```
# gdk:myorg/private-repo - Added by gdk on 2025-01-15
Host github.com-myorg-private-repo
    HostName github.com
    User git
    IdentityFile ~/.ssh/deploy-keys/myorg-private-repo/id_ed25519
    IdentitiesOnly yes
```

4. **Clone URL**: Use the host alias in your git commands:
```bash
git clone git@github.com-myorg-private-repo:myorg/private-repo.git
```

## File Structure

```
~/.ssh/
├── config                           # SSH config (modified by gdk)
└── deploy-keys/                     # Created by gdk
    ├── .gdk-metadata.json          # Tracks all managed keys
    └── myorg-private-repo/
        ├── id_ed25519              # Private key (0600)
        └── id_ed25519.pub          # Public key (0644)
```

## CI/CD

This project uses GitHub Actions for CI/CD:

- **CI**: Runs tests on every push to `master` and on pull requests
- **Release**: Builds binaries for Linux, macOS, and Windows when a tag is pushed

### Creating a Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

This triggers the release workflow which:
1. Runs all tests
2. Builds binaries for all platforms
3. Creates a GitHub release with downloadable binaries

## Development

### Prerequisites

- Go 1.21 or later

### Building

```bash
go build -o gdk .
```

### Running Tests

```bash
go test ./... -v
```

### Test Coverage

```bash
go test ./... -cover
```

## License

MIT License
