# CLAUDE.md - Development Guidelines for codacy-cli-v2

## Build & Test Commands
- Build: `go build ./cli-v2.go`
- Run all tests: `go test ./...`
- Run specific test: `go test -run TestName ./package/path`
- Example: `go test -run TestGetTools ./tools`
- Format code: `go fmt ./...`
- Lint: `golint ./...` (if installed)

## Code Style Guidelines
- **Imports**: Standard lib first, external packages second, internal last
- **Naming**: PascalCase for exported (public), camelCase for unexported (private)
- **Error handling**: Return errors as last value, check with `if err != nil`
- **Testing**: Use testify/assert package for assertions
- **Package organization**: Keep related functionality in dedicated packages
- **Documentation**: Document all exported functions, types, and packages
- **Commit messages**: Start with verb, be concise and descriptive

## Project Structure
- `cmd/`: CLI command implementations
- `config/`: Configuration handling
- `tools/`: Tool-specific implementations
- `utils/`: Utility functions