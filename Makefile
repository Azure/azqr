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
	@echo "  $(TARGET)    - Build the azqr binary (production)"
	@echo "  debug        - Build the azqr binary with profiling support"
	@echo "  lint         - Run linting checks"
	@echo "  vet          - Run go vet checks"
	@echo "  tidy         - Tidy up go modules and check for changes"
	@echo "  json         - Generate JSON recommendations and check for changes"
	@echo "  test         - Run tests (includes linting)"
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

json:
	go run ./cmd/azqr/main.go rules --json > ./data/recommendations.json 
	git diff --exit-code ./data/recommendations.json

test: lint vet tidy json
	go test -race ./... -coverprofile=coverage.txt -covermode=atomic ./...

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
