.DEFAULT_GOAL := all

TARGET     := azqr
OS         := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH       := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOARM      := $(if $(GOARM),$(GOARM),)
BIN         = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET)
ifeq ($(OS),windows)
  BIN = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET).exe
endif
GOLANGCI_LINT := ./bin/golangci-lint

PRODUCT_VERSION	:= $(if $(PRODUCT_VERSION),$(PRODUCT_VERSION),'dev')

# Build flags for better antivirus compatibility and Windows Defender ASR rules
# Carefully chosen flags to minimize false positives while maintaining functionality
ifeq ($(GOOS),windows)
  # For Windows, use minimal stripping and preserve build metadata for better reputation
  # Avoid removing all debug information to reduce ASR rule triggers
  LDFLAGS	:= -X github.com/Azure/azqr/cmd/azqr/commands.version=$(PRODUCT_VERSION) -extldflags="-static"
  # Add build tags for Windows compatibility and security
  BUILD_TAGS := -tags="netgo,osusergo" -buildmode=exe
  # Add trimpath to remove local file system paths from binary for better security
  TRIM_PATH := -trimpath
else
  # For other platforms, use full stripping for smaller binaries
  LDFLAGS	:= -s -w -X github.com/Azure/azqr/cmd/azqr/commands.version=$(PRODUCT_VERSION)
  BUILD_TAGS := -tags="netgo"
  TRIM_PATH := -trimpath
endif

all: $(TARGET)

help:
	@echo "Available targets:"
	@echo "  all          - Build the azqr binary (default)"
	@echo "  $(TARGET)    - Build the azqr binary"
	@echo "  lint         - Run linting checks"
	@echo "  vet          - Run go vet checks"
	@echo "  tidy         - Tidy up go modules and check for changes"
	@echo "  json         - Generate JSON recommendations and check for changes"
	@echo "  test         - Run tests (includes linting)"
	@echo "  clean        - Remove built binaries"
	@echo "  build-image  - Build Docker image with azqr binary"
	@echo ""
	@echo "Docker image build options:"
	@echo "  make build-image                      # Build with 'latest' tag"
	@echo "  PRODUCT_VERSION=1.0.0 make build-image   # Build with specific tag"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS         - Target OS (default: $(OS))"
	@echo "  GOARCH       - Target architecture (default: $(ARCH))"
	@echo "  PRODUCT_VERSION - Git tag for version info and Docker tagging"

lint:
	@if [ ! -f $(GOLANGCI_LINT) ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.3.0; \
	fi
	$(GOLANGCI_LINT) run

vet:
	go vet ./...

tidy:
	go mod tidy
	git diff --exit-code ./go.mod
	git diff --exit-code ./go.sum

test: lint vet tidy
	go test -race ./... -coverprofile=coverage.txt -covermode=atomic ./...

$(TARGET): clean
	CGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0) go build $(TRIM_PATH) $(BUILD_TAGS) -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/azqr/main.go

clean:
	-rm -f $(BIN)

json:
	go run ./cmd/azqr/main.go rules --json > ./data/recommendations.json 
	git diff --exit-code ./data/recommendations.json

# Docker image build target
IMAGE_NAME    := ghcr.io/azure/azqr
IMAGE_TAG     := $(if $(PRODUCT_VERSION),$(PRODUCT_VERSION),latest)

build-image: $(TARGET)
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@if [ "$(PRODUCT_VERSION)" != "" ]; then \
		docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest; \
	fi