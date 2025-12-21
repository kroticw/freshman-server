.PHONY: test
test:
	@go test -v ./...

.PHONY: serve
serve: ## Запускает в режиме Serve
	@go run . serve --config=config.yaml
