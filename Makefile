.PHONY: build run generate css migrate-up migrate-down clean install-tools

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the development server
run: css
	go run ./cmd/server

# Generate sqlc code
generate: sqlc

# Generate sqlc code
sqlc:
	sqlc generate

# Generate templ templates
templ:
	templ generate

# Build CSS
css:
	npm run build:css

# Run migrations up
migrate-up:
	goose -dir migrations sqlite3 ./data.db up

# Run migrations down
migrate-down:
	goose -dir migrations sqlite3 ./data.db down

# Check migration status
migrate-status:
	goose -dir migrations sqlite3 ./data.db status

# Clean generated files
clean:
	rm -rf bin/
	rm -f internal/db/*.go
	rm -f internal/templates/*.go
	rm -f static/output.css

# Install development tools
install-tools:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest