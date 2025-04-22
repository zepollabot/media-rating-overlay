#SHELL := /bin/sh

# import env variables
# you can change the default filename with `make env_file="config_special.env" build`
env_file ?= .env
ifneq ("$(wildcard $(env_file))", "")
	include $(env_file)
	export $(shell sed 's/=.*//' $(env_file))
endif

# Variables for Docker image and registry
IMAGE_NAME ?= media-rating-overlay
# Attempt to derive GITHUB_OWNER from git remote, can be overridden
GITHUB_OWNER ?= $(shell git remote get-url origin 2>/dev/null | sed -n 's|.*/\\([^/]*\\)/[^/]*\\.git|\\1|p')
GHCR_IMAGE_PATH := ghcr.io/$(GITHUB_OWNER)/$(IMAGE_NAME)

TPUT := $(shell [ -n "$$TERM" ] && [ "$$TERM" != "dumb" ] && echo 'tput' || echo 'tput -Tvt100')

BOLD := "$$([ -t 0 ] && $(TPUT) bold)"
NORMAL := "$$([ -t 0 ] && $(TPUT) sgr0)"
GREEN := "$$([ -t 0 ] && $(TPUT) setaf 2)"

ABSOLUTE_PROJECT_PATH := $(shell git rev-parse --show-toplevel)

$(VERBOSE).SILENT:

.PHONY: help
help: ## Will print this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


.PHONY: setup-config
setup-config: ## Setup the project
	cp configs/config.yaml.dist configs/config.yaml; \
	cp configs/config.env.prod.yaml.dist configs/config.env.prod.yaml

.PHONY: run
run: ## Run the project
	go mod download; \
	go run main.go

.PHONY: run-docker
run-docker: ## Run the project using Docker compose
	docker compose up

.PHONY: build-dev
build-dev: ## Build the project for DEV env using Docker compose
	MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose build app

.PHONY: run-dev
run-dev: ## Run the project for DEV env using Docker compose
	docker compose run --rm app;

.PHONY: build-and-run-dev
build-and-run-dev: ## Build and run the project for DEV env using Docker compose
	make build-dev; \
	make run-dev

.PHONY: build-prod
build-prod: ## Build a Docker image for PROD env
	docker build --target prod -t media-rating-overlay -f docker/golang/Dockerfile .

.PHONY: run-prod
run-prod: ## Run the PROD Docker image
	docker run -v $(shell pwd)/configs:/configs -v /Multimedia:/Multimedia -u $(id -u ${USER}):$(id -g ${USER}) --rm media-rating-overlay;

.PHONY: build-and-run-prod
build-and-run-prod: ## Build and run the PROD Docker image
	make build-prod; \
	make run-prod

.PHONY: pack-prod
pack-prod: ## Create docker image tar.gz
	make build-prod; \
	docker save media-rating-overlay:latest | gzip > media_rating_overlay_latest.tar.gz

.PHONY: release-version
release-version: ## Tag and push a new version (e.g., make release-version VERSION=v1.0.1)
ifndef VERSION
	$(error VERSION is not set. Usage: make release-version VERSION=vX.Y.Z)
endif
	@echo "$(GREEN)Tagging version $(VERSION)...$(NORMAL)"
	git tag $(VERSION)
	@echo "$(GREEN)Pushing tag $(VERSION) to origin...$(NORMAL)"
	git push origin $(VERSION)
	@echo "$(GREEN)GitHub Action should now build and release image: $(GHCR_IMAGE_PATH):$(VERSION)$(NORMAL)"

.PHONY: download-image
download-image: ## Download a tagged image from GHCR and save as tar.gz (e.g., make download-image TAG=v1.0.1 or TAG=latest)
ifndef TAG
	$(error TAG is not set. Usage: make download-image TAG=vX.Y.Z or TAG=latest)
endif
ifndef GITHUB_OWNER
	$(error GITHUB_OWNER is not set. Please set it in your .env file or pass it as an argument)
endif
	@echo "$(GREEN)Pulling image $(GHCR_IMAGE_PATH):$(TAG)...$(NORMAL)"
	docker pull $(GHCR_IMAGE_PATH):$(TAG)
	@echo "$(GREEN)Saving image $(GHCR_IMAGE_PATH):$(TAG) to $(IMAGE_NAME)_$(TAG).tar.gz...$(NORMAL)"
	docker save $(GHCR_IMAGE_PATH):$(TAG) | gzip > $(IMAGE_NAME)_$(TAG).tar.gz
	@echo "$(GREEN)Image saved as $(IMAGE_NAME)_$(TAG).tar.gz$(NORMAL)"

.PHONY: mocks
mocks: ## Generate mocks
	MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose run --rm app mockery --config .mockery.yaml

.PHONY: clean-mocks
clean-mocks: ## Clean all the generated mocks
	find . -type d -name mocks -exec sudo rm -rf {} +

.PHONY: test
test: mocks ## Run tests
	MY_UID="$(id -u)" MY_GID="$(id -g)" docker compose run --rm app sh -c "go list ./... | grep -v -E '(/vendor/|/mocks)' | xargs go test -v -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html"; \

.PHONY: test-coverage-view
test-coverage-view: ## Open the coverage report in the default browser
	xdg-open coverage.html || open coverage.html || echo "Please open coverage.html in your browser"

.PHONY: test-coverage-clean
test-coverage-clean: ## Clean the coverage report
	rm -f coverage.out coverage.html
