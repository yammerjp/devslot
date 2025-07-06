# devslot

<p align="center">
  <img src="assets/logo.jpg" alt="devslot logo" width="600">
</p>

<p align="center">
  A development environment manager for multi-repository projects using Git worktrees.
</p>

## Overview

devslot helps you manage multiple Git repositories as a cohesive development environment. It creates isolated "slots" where each repository is checked out as a Git worktree, allowing you to work on different features or branches across multiple repositories simultaneously.

## Features

- **Multi-repository management**: Work with multiple related repositories as a single unit
- **Isolated environments**: Create separate slots for different features or experiments
- **Git worktree based**: Leverages Git's worktree feature for efficient disk usage
- **Lifecycle hooks**: Automate setup and teardown with customizable scripts
- **Branch management**: Automatically creates consistent branch names across repositories

## Installation

### Using Go

```bash
go install github.com/yammerjp/devslot/cmd/devslot@latest
```

### From source

```bash
git clone https://github.com/yammerjp/devslot.git
cd devslot
make build
# Binary will be at ./build/devslot
```

## Quick Start

1. **Initialize a new project**:
   ```bash
   devslot boilerplate my-project
   cd my-project
   ```

2. **Configure your repositories** in `devslot.yaml`:
   ```yaml
   version: 1
   repositories:
     - name: frontend
       url: https://github.com/myorg/frontend.git
     - name: backend
       url: https://github.com/myorg/backend.git
   ```

3. **Clone the repositories**:
   ```bash
   devslot init
   ```

4. **Create a development slot**:
   ```bash
   devslot create feature-x
   ```

This creates a new slot with worktrees for all repositories, ready for development.

## Commands

- `devslot boilerplate <dir>` - Generate initial project structure
- `devslot init` - Clone repositories defined in devslot.yaml
- `devslot create <slot>` - Create a new development slot
- `devslot list` - List all existing slots
- `devslot destroy <slot>` - Remove a slot
- `devslot reload <slot>` - Synchronize slot with current configuration
- `devslot doctor` - Check project health
- `devslot version` - Show version information

Run `devslot <command> --help` for detailed information about each command.

## Configuration

### devslot.yaml

The project configuration file that defines your repositories:

```yaml
version: 1
repositories:
  - name: app
    url: https://github.com/example/app.git
  - name: lib
    url: https://github.com/example/lib.git
```

### Hooks

Optional lifecycle scripts in the `hooks/` directory:

- `post-init` - Runs after `devslot init`
- `post-create` - Runs after creating a slot
- `pre-destroy` - Runs before destroying a slot
- `post-reload` - Runs after reloading a slot

Hooks receive environment variables with context about the operation. See the generated examples for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Release Process

This project uses [tagpr](https://github.com/Songmu/tagpr) for automated releases. When changes are merged to main, tagpr automatically creates a release PR. To release:

- Merge the auto-generated PR to create a new release
- For minor version bumps: add `tagpr:minor` label to the PR
- For major version bumps: add `tagpr:major` label to the PR

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.