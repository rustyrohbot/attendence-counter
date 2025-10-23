# Agent Guidelines for attendance-tracker

## Build Commands
- `make build` - Build the application binary
- `make run` - Run the development server
- `make generate` - Generate sqlc and templ code
- `make css` - Build CSS with Tailwind
- `make migrate-up` - Run database migrations

## Test Commands
- No tests currently exist - run `go test ./...` after adding tests
- For single test: `go test -run TestName ./package/path`

## Code Style Guidelines
- **Go Version**: 1.21+
- **Formatting**: Use standard `gofmt`
- **Imports**: Standard library → third-party → local (attendance-tracker/...)
- **Naming**: PascalCase for exported, camelCase for unexported
- **Error Handling**: Check errors immediately, use `http.Error()` in handlers
- **Types**: Use `sql.NullString`/`sql.NullInt64` for nullable DB fields
- **Structs**: Use pointer receivers for methods
- **Database**: Use sqlc-generated queries, pass context to all DB operations
- **Templates**: Use templ for type-safe HTML generation
- **HTTP**: Use standard `net/http`, handle forms with `r.ParseForm()`
- **Dates**: Store as TEXT in YYYY-MM-DD format, display as readable text