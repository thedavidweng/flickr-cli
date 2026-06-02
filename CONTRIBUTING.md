# Contributing

## Getting Started

```bash
git clone https://github.com/thedavidweng/flickr-cli.git
cd flickr-cli
go build ./cmd/flickr
go test ./...
```

Requires Go 1.26.3 or later.

## Development

```bash
# Build
make build

# Run tests
make test

# Format, vet, and test
make lint

# Format only
make fmt

# Vet only
make vet
```

## Project Structure

```
cmd/flickr/           Entry point
internal/
  cli/                Cobra commands
  flickr/             Flickr API client
  config/             Configuration
  model/              Domain types
  output/             Rendering
  backup/             Backup logic
  upload/             Upload logic
  cache/              SQLite cache
  checksum/           Hash and machine tags
  safety/             Safety gates and audit
  piwigo/             Piwigo import
  testutil/           Test helpers
docs/                 Documentation
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture.

## Code Style

- Standard Go formatting (`gofmt`)
- Table-driven tests
- No CGO dependencies
- Errors wrapped with `fmt.Errorf("context: %w", err)`
- Secrets never printed in any output mode

## Testing

```bash
# All tests
go test ./...

# Specific package
go test ./internal/cli/

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
```

The fake Flickr server (`internal/testutil/`) provides a complete mock for integration tests without hitting the real API.

## Pull Requests

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure `make lint` passes
5. Submit a pull request

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
