# project-templating

A CLI tool that processes template files and generates output by substituting variables. It automatically provides Git-aware and color-themed variables from your working environment, and accepts additional custom variables via config file or CLI arguments.

Templates use Go's `text/template` syntax: `{{ .VARIABLE_NAME }}`.

## Usage

```bash
./project-cli [OPTIONS] [KEY=VALUE ...]
```

**The current directory must be inside a Git repository.**

### Options

| Flag      | Default     | Description                                   |
| --------- | ----------- | --------------------------------------------- |
| `-src`    | `templates` | Source directory containing template files    |
| `-dest`   | `.`         | Destination directory for generated output    |
| `-config` | _(none)_    | Path to a JSON file with additional variables |

Custom variables can also be passed as trailing `KEY=VALUE` arguments.

### Examples

```bash
# Process templates/ and output to current directory
./project-cli

# Custom source and destination
./project-cli -src ./my-templates -dest ./output

# Load extra variables from a JSON file
./project-cli -config vars.json

# Pass variables directly on the command line
./project-cli APP_NAME=my-service ENV=production

# All combined
./project-cli -src templates -dest ./gen -config config.json APP_NAME=myapp ENV=production
```

## Built-in Variables

These are always available in every template run — no configuration needed.

| Variable                            | Type   | Description                                                                                |
| ----------------------------------- | ------ | ------------------------------------------------------------------------------------------ |
| `BG_COLOR`                          | string | Hex color (`#RRGGBB`) derived deterministically from the current working directory path    |
| `FG_COLOR`                          | string | Contrasting foreground color (`#000000` or `#ffffff`) calculated from `BG_COLOR` luminance |
| `IS_GIT_WORKTREE`                   | bool   | `true` if the repo is a Git worktree; `false` if it's the primary working tree             |
| `GIT_REL_SOURCE`                    | string | Relative path from the current directory to the Git directory (e.g., `../.git`)            |
| `GIT_ABS_TARGET`                    | string | Absolute path to the Git directory (e.g., `/path/to/repo/.git`)                            |
| `GIT_BRANCH`                        | string | Current branch name, or short commit hash in detached HEAD state                           |
| `GIT_WORKTREE_MAIN_FOLDER_PATH`     | string | Absolute path to the main worktree folder (the one containing `.git`)                      |
| `GIT_WORKTREE_MAIN_FOLDER_BASENAME` | string | Folder name of the main worktree (e.g., `my-repo`)                                         |

## Custom Variables

Variables are merged from three sources, in order of increasing precedence:

1. **Built-in variables** (always present)
2. **JSON config file** (`-config` flag)
3. **CLI arguments** (`KEY=VALUE` trailing args — highest precedence)

### JSON Config File Format

```json
{
  "APP_NAME": "my-service",
  "DATABASE_URL": "postgresql://localhost/mydb",
  "ENVIRONMENT": "production"
}
```

All keys in the JSON object become available as template variables.

## Template Syntax

Templates use Go's [`text/template`](https://pkg.go.dev/text/template) package. Every file in the source directory is processed as a template — all file types are supported.

```
# {{ .APP_NAME }}

Branch: {{ .GIT_BRANCH }}
Is worktree: {{ .IS_GIT_WORKTREE }}
Background color: {{ .BG_COLOR }}

{{ if .IS_GIT_WORKTREE }}
Git dir: {{ .GIT_REL_SOURCE }}
{{ end }}
```

Directory structure is preserved: a file at `templates/subdir/file.txt` is written to `<dest>/subdir/file.txt`. Parent directories are created automatically.

## Installation

Pre-built binaries for Darwin, Linux, and Windows (amd64 and arm64) are in [bin/](bin/). The `project-cli` wrapper script selects the correct binary for the current platform automatically.

To build from source:

```bash
./scripts/build.sh
```

## Testing

```bash
./scripts/test.sh
```
