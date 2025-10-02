PKGS := ./...
COVER_FILE := coverage.out

# Default target
.PHONY: test
test: ## Run tests with coverage
	@echo "Running tests..."
	@. ./export.sh && go test -v -coverprofile=$(COVER_FILE) $(PKGS)
	@echo
	@echo "Coverage summary:"
	go tool cover -func=$(COVER_FILE)

.PHONY: cover
cover: test ## Show HTML coverage report
	@echo "Opening HTML coverage report..."
	go tool cover -html=$(COVER_FILE)

.PHONY: clean
clean: ## Clean up coverage files
	rm -f $(COVER_FILE)

.PHONY: dc-test
dc-test: ## Run tests inside docker-compose
	@echo "Running tests in Docker..."
	@docker compose run --rm app

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'


%:
	@:
