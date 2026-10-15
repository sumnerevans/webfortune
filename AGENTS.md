# Repository Guidelines

## Project Structure & Module Organization

`main.go` configures the HTTP server and routes. `quotesfile/` parses fortune-style quote files and models quotes. `templates/` contains templ components (`*.templ`) and their Go parameters; generated `*_templ.go` files are ignored. `screenshot.png` is the README image. There is currently no checked-in test suite or bundled quotes file.

## Build, Test, and Development Commands

Use Go 1.26 (or `nix develop` on x86_64 Linux). Run `go generate ./...` after editing templates; it invokes `go tool templ generate`. Run `go build` to compile the server, or `nix build` to build the flake package. Start a local instance with a valid fortune-style file:

```sh
QUOTESFILE=/path/to/quotes LISTEN_ADDR=:8080 HOST_ROOT=http://localhost:8080 go run .
```

Run `go test ./...` for tests and `pre-commit run --all-files` for the configured formatting, static analysis, vet, and logging checks. CI runs generation, pre-commit, and `go build` on pushes.

## Coding Style & Naming Conventions

Follow `.editorconfig`: tabs for Go, two-space indentation for Markdown, YAML, and Nix, LF endings, and a final newline. Use `goimports` for Go formatting and import grouping; the pre-commit hook runs it with this module as the local prefix. Keep template changes in `*.templ` rather than editing generated Go. Use idiomatic Go names: exported identifiers in `PascalCase`, unexported identifiers in `camelCase`, and test files named `*_test.go`.

## Testing Guidelines

The repository has no existing tests or coverage threshold. Add focused Go tests alongside changed packages, especially for quote parsing, hash links, and HTTP handlers. Name tests `TestXxx`, use representative fortune-style input (quotes separated by a line containing `%`), and run `go test ./...` before submitting.

## Commit & Pull Request Guidelines

Follow [Scoped Commits](https://scopedcommits.com/): write subjects as `<scope>: <short description>`, where the scope names the affected area. Recent examples use `templates:`, `all:`, `ci:`, and `deps:`; dependency automation also uses `build(deps):`. Keep commits focused. In pull requests, describe the behavior changed, note validation commands, link a relevant issue when one exists, and include a screenshot for visible UI changes. Document any new environment variables; current settings are listed in `README.md`.
