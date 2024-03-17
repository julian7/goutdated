# goutdated

This is a short and sweet utility to update `go.mod` dependencies.

## Usage

`goutdated [-a] [-n]`

Options:

- `-a`: Show all. It shows all outdated packages used with current versions. Not more sophisticated than `go list -u -m all | grep -F '['`.
- `-n`: Dry run. It shows all outdated packages referenced by `go.mod` file, hinting on version changes.

Running it with no arguments rewrites `go.mod` to use latest versions. It's always a good idea to run `go mod tidy` before/after `goutdated`, and re-run `goutdated` until there are no modifications. This way the tool can pick up changed dependencies too.
