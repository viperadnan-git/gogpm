# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Rules

- Always implement root fixes and never add patch fixes
- Update this doc if any reference is changed
- Write minimal lines of code

## Build Commands

```bash
make build      # Build to ./cmd/gpcli/gpcli
make clean      # Remove built binary
cd cmd/gpcli && go build -o ../../gpcli .   # Direct build command (not recommended)
```

## Protobuf Generation

See `.proto/README.md` for detailed instructions on generating protobuf files.

Quick reference from project root:
```bash
export PATH=$PATH:$(go env GOPATH)/bin

# Generate single file
protoc --proto_path=. --go_out=. --go_opt=module=github.com/viperadnan-git/go-gpm .proto/MessageName.proto

# Generate all files
for proto in .proto/*.proto; do
  protoc --proto_path=. --go_out=. --go_opt=module=github.com/viperadnan-git/go-gpm "$proto"
done
```

## Architecture

This is a monorepo containing both a CLI tool and a Go library for managing Google Photos using an unofficial API. It uses protobuf for API communication.

### Module Structure

The project uses two Go modules to separate library and CLI dependencies:

- **Root module** (`go.mod`): `github.com/viperadnan-git/go-gpm` - Library with minimal dependencies (protobuf, retryablehttp)
- **CLI module** (`cmd/gpcli/go.mod`): `github.com/viperadnan-git/go-gpm/cmd/gpcli` - CLI with additional dependencies (urfave/cli, koanf, go-selfupdate)

The CLI module uses a `replace` directive for local development:
```go
replace github.com/viperadnan-git/go-gpm => ../..
```

### Key Components

- **Root package (gpm)** - Public library API
  - `main.go` - GooglePhotosAPI struct embedding internal/core.Api
  - `uploader.go` - Upload orchestration with worker pool. Emits progress events.
  - `utils.go` - Download utilities, ResolveItemKey, ResolveMediaKey
  - `sha1.go` - File hash calculation
  - `version.go` - Version constant
- **cmd/gpcli/** - CLI application using urfave/cli/v3
  - `main.go` - Entry point + command definitions for upload, download, thumbnail, auth, delete, archive, favourite, caption, upgrade
  - `config.go` - YAML config file management. Stores credentials and settings.
- **internal/core/** - Low-level API operations (not exported)
  - `api.go` - Api struct with auth token management and common headers
  - `upload.go` - Upload token, file upload, commit operations
  - `download.go` - Download URL retrieval
  - `trash.go` - MoveToTrash, RestoreFromTrash operations
  - `archive.go` - SetArchived operation
  - `metadata.go` - SetCaption, SetFavourite operations
  - `album.go` - CreateAlbum, AddMediaToAlbum operations
  - `thumbnail.go` - Thumbnail download
  - `utils.go` - SHA1ToDedupeKey, ToURLSafeBase64
- **internal/pb/** - Protobuf-generated Go code for API request/response structures (not exported)
- **.proto/** - Protobuf definitions for Google Photos API messages

### Event-Based Progress System

The upload system uses an event callback pattern:
1. `GooglePhotosAPI.Upload()` starts worker goroutines
2. Workers emit events via callback function
3. CLI receives events and prints progress to stdout

Event types: `uploadStart`, `ThreadStatus`, `FileStatus`, `uploadStop`

### Config File

Config is stored in `./gpcli.config` (YAML) or custom path via `--config` flag. Contains credentials array and upload settings.
