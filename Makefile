.DEFAULT_GOAL := all

TARGET     := azqr
OS         := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH       := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOARM      := $(if $(GOARM),$(GOARM),)
BIN         = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET)
DEBUG_BIN   = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET)-debug
ifeq ($(OS),windows)
  BIN = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET).exe
  DEBUG_BIN = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET)-debug.exe
endif
GOLANGCI_LINT := ./bin/golangci-lint
PRODUCT_VERSION	:= $(if $(PRODUCT_VERSION),$(PRODUCT_VERSION),'0.0.0-dev')
LDFLAGS	:= -s -w -X github.com/Azure/azqr/cmd/azqr/commands.version=$(PRODUCT_VERSION)
DEBUG_LDFLAGS := -X github.com/Azure/azqr/cmd/azqr/commands.version=$(PRODUCT_VERSION)-debug
TRIM_PATH := -trimpath
BUILD_TAGS := $(if $(BUILD_TAGS),$(BUILD_TAGS),)
DEBUG_BUILD_TAGS := debug

all: $(TARGET)

build: $(TARGET)

debug: $(TARGET)-debug

help:
	@echo "Available targets:"
	@echo "  all          - Build the azqr binary (default)"
	@echo "  build        - Build the azqr binary (same as all)"
	@echo "  $(TARGET)    - Build the azqr binary (production)"
	@echo "  debug        - Build the azqr binary with profiling support"
	@echo "  lint         - Run linting checks"
	@echo "  lint-all     - Run comprehensive linting checks (includes errcheck, gosec, etc.)"
	@echo "  vet          - Run go vet checks"
	@echo "  tidy         - Tidy up go modules and check for changes"
	@echo "  json         - Generate JSON recommendations and check for changes"
	@echo "  validate-yaml - Validate all recommendation YAML files against schema"
	@echo "  validate-scanners - Validate APRL recommendations coverage"
	@echo "  test         - Run tests (includes linting and validation)"
	@echo "  test-integration - Run integration tests (requires Azure credentials)"
	@echo "  test-integration-setup - Validate prerequisites for running integration tests"
	@echo "  terraform-fmt - Format all Terraform files"
	@echo "  terraform-validate - Validate all Terraform fixtures"
	@echo "  clean        - Remove built binaries"
	@echo "  build-image  - Build Docker image with azqr binary"
	@echo ""
	@echo "Build options:"
	@echo "  make debug                               # Build with profiling support"
	@echo "  make $(TARGET)                          # Build production version (no profiling)"
	@echo ""
	@echo "Docker image build options:"
	@echo "  make build-image                         # Build with 'latest' tag"
	@echo "  PRODUCT_VERSION=1.0.0 make build-image   # Build with specific tag"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS         - Target OS (default: $(OS))"
	@echo "  GOARCH       - Target architecture (default: $(ARCH))"
	@echo "  PRODUCT_VERSION - Git tag for version info and Docker tagging"
	@echo "  BUILD_TAGS   - Custom build tags to include"

lint: lint-install
	$(GOLANGCI_LINT) run

lint-all: lint-install
	$(GOLANGCI_LINT) run --enable=errcheck,govet,ineffassign,staticcheck,gocyclo,unused,gocognit,gosec,gocritic

lint-install:
	@if [ ! -f $(GOLANGCI_LINT) ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.9.0; \
	fi

vet:
	go vet ./...

tidy:
	go mod tidy
	git diff --exit-code ./go.mod
	git diff --exit-code ./go.sum

json:
	go run ./cmd/azqr/main.go rules --json > ./data/recommendations.json 
	git diff --exit-code ./data/recommendations.json

validate-yaml:
	@echo "Validating recommendation YAML files against schema..."
	@go run ./scripts/validate-recommendations.go ./internal/graph/azqr/azure-resources ./internal/graph/aprl/azure-resources ./internal/graph/azure-orphan-resources

validate-scanners: validate-yaml
	@./scripts/validate-scanner-coverage.sh

test: lint vet tidy json validate-yaml validate-scanners
	go test -race ./... -coverprofile=coverage.txt -covermode=atomic ./...

# Integration test targets
test-integration-setup:
	@echo "Checking integration test prerequisites..."
	@if [ -z "$$AZURE_SUBSCRIPTION_ID" ]; then \
		echo "❌ AZURE_SUBSCRIPTION_ID environment variable is not set"; \
		exit 1; \
	fi
	@if [ -z "$$AZURE_TENANT_ID" ]; then \
		echo "❌ AZURE_TENANT_ID environment variable is not set"; \
		exit 1; \
	fi
	@echo "✓ AZURE_SUBSCRIPTION_ID is set: $$AZURE_SUBSCRIPTION_ID"
	@echo "✓ AZURE_TENANT_ID is set: $$AZURE_TENANT_ID"
	@command -v terraform >/dev/null 2>&1 || { echo "❌ terraform is not installed or not in PATH"; exit 1; }
	@echo "✓ terraform is installed: $$(terraform version -json 2>/dev/null | grep -oP '(?<="version":")[^"]*' || terraform version | head -1)"
	@echo "✓ Prerequisites check passed!"

test-integration: test-integration-setup
	@echo "Cleaning terraform state and lock files..."
	@find ./test/fixtures/terraform -name "terraform.tfstate*" -delete 2>/dev/null || true
	@find ./test/fixtures/terraform -name ".terraform.lock.hcl" -delete 2>/dev/null || true
	@echo "✓ Terraform state cleaned"
	@echo "Running integration tests..."
	@echo "⚠ This will provision real Azure resources and may incur costs"
	@echo "📝 Terraform output will be shown during test execution"
	go test -v -p 1 -tags=integration -timeout 30m ./test/integration/... 2>&1

terraform-fmt:
	@echo "Formatting Terraform files..."
	@find ./test/fixtures/terraform -name "*.tf" -exec terraform fmt {} \;
	@echo "✓ Terraform files formatted"

terraform-validate:
	@echo "Validating Terraform fixtures..."
	@for dir in $$(find ./test/fixtures/terraform -type d -name "baseline" -o -name "scenarios" | xargs -I {} find {} -mindepth 1 -maxdepth 1 -type d); do \
		echo "Validating $$dir..."; \
		(cd "$$dir" && terraform init -backend=false >/dev/null 2>&1 && terraform validate) || exit 1; \
	done
	@echo "✓ All Terraform fixtures validated"


$(TARGET): clean
	CGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0) go build $(TRIM_PATH) -tags "$(BUILD_TAGS)" -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/azqr/main.go

$(TARGET)-debug: clean
	CGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0) go build $(TRIM_PATH) -tags "$(DEBUG_BUILD_TAGS)" -o $(DEBUG_BIN) -ldflags "$(DEBUG_LDFLAGS)" ./cmd/azqr/main.go
	@echo "Debug build created at: $(DEBUG_BIN)"
	@echo "This build includes profiling support with flags: --cpu-profile, --mem-profile, --trace-profile"

clean:
	-rm -f $(BIN)
	-rm -f $(DEBUG_BIN)

# Docker image build target
IMAGE_NAME    := ghcr.io/azure/azqr
IMAGE_TAG     := $(if $(PRODUCT_VERSION),$(PRODUCT_VERSION),latest)

build-image: $(TARGET)
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@if [ "$(PRODUCT_VERSION)" != "" ]; then \
		docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest; \
	fi
