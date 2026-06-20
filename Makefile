.PHONY: docs build-site test-install

docs: ## Regenerate the committed command reference
	go run ./cmd/gen-docs

build-site: ## Build the website
	cd website && npm ci && npm run build

test-install: ## Unit-test the install script (no network)
	sh test/install_test.sh
