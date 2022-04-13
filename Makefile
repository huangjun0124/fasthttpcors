GO=go
GOTOOL=go tool cover
MOCKGEN=mockgen

vet: ## Runs govet against all packages.
	@echo Running GOVET
	$(shell go vet $(GOFLAGS) ./...)

clean: ## Clean up everything except persistant server data.
	@echo Cleaning
	go clean $(GOFLAGS) -i ./...
	go mod tidy

ver = 'v1.0.0'
release:
	npm run release -- --release-as $(ver)
	# git push --follow-tags origin master