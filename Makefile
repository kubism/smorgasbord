SHELL := bash

# Directory, where all required tools are located (absolute path required)
TOOLS_DIR ?= $(shell cd tools && pwd)
# Assets for dex to successfully run tests
DEX_WEB_DIR ?= $(shell cd web && pwd)

# Prerequisite tools
GO ?= go
DOCKER ?= docker

# Tools managed by this project
GINKGO ?= $(TOOLS_DIR)/ginkgo
LINTER ?= $(TOOLS_DIR)/golangci-lint
GOVERALLS ?= $(TOOLS_DIR)/goveralls
GOVER ?= $(TOOLS_DIR)/gover
2GOARRAY ?= $(TOOLS_DIR)/2goarray

# Variables relevant for builds
BUILD_PATH ?= $(shell pwd)
VERSION    ?= 0.0.0-dev
COMMIT     := $(shell git rev-parse --short HEAD)
LDFLAGS    += -ldflags "-X=main.version=$(VERSION) -X=main.commit=$(COMMIT)"
BUILDFLAGS += -installsuffix cgo --tags release

ICON_UNIX_SRC ?= assets/icon/icon.png
ICON_UNIX ?= assets/icon/iconunix.go
ICON_WIN_SRC ?= assets/icon/iconwin.ico
ICON_WIN ?= assets/icon/iconwin.go

# Binaries to build
CMD_SMORGASBORD = $(BUILD_PATH)/smorgasbord
CMD_SMORGASBORD_SRC = cmd/smorgasbord/*.go

CMD_UITEST = $(BUILD_PATH)/uitest
CMD_UITEST_SRC = cmd/uitest/*.go


.EXPORT_ALL_VARIABLES:
.PHONY: build clean test lint fmt vet icon

$(CMD_SMORGASBORD): $(CMD_SMORGASBORD_SRC)
	CGO_ENABLED=0 $(GO) build -o $(CMD_SMORGASBORD) -a $(BUILDFLAGS) $(LDFLAGS) $(CMD_SMORGASBORD_SRC)

$(CMD_UITEST): $(CMD_UITEST_SRC) icon
	$(GO) build -o $(CMD_UITEST) -a $(BUILDFLAGS) $(LDFLAGS) $(CMD_UITEST_SRC)

run-%: icon
	$(GO) run cmd/$*/*.go


build: $(CMD_SMORGASBORD)

clean:
	rm -f $(CMD_SMORGASBORD)

test: fmt vet $(GINKGO)
	$(GINKGO) -r -v -cover . -- -dex-web-dir=$(DEX_WEB_DIR) $(TEST_FLAGS)


test-%: fmt vet $(GINKGO)
	$(GINKGO) -r -v -cover pkg/$* -- -dex-web-dir=$(DEX_WEB_DIR) $(TEST_FLAGS)

# First run gover to merge the coverprofiles and upload to coveralls
coverage: $(GOVERALLS) $(GOVER)
	$(GOVER)
	$(GOVERALLS) -coverprofile=gover.coverprofile -service=travis-ci -repotoken $(COVERALLS_TOKEN)

lint: $(LINTER)
	$(GO) mod verify
	$(LINTER) run -v --no-config --deadline=5m

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

icon: $(ICON_UNIX) $(ICON_WIN)

$(ICON_UNIX): $(2GOARRAY)
	echo -e "//+build linux darwin" > $(ICON_UNIX)
	cat $(ICON_UNIX_SRC) | $(2GOARRAY) Data icon >> $(ICON_UNIX)

$(ICON_WIN): $(2GOARRAY)
	echo -e "//+build windows" > $(ICON_WIN)
	cat $(ICON_WIN_SRC) | $(2GOARRAY) Data icon >> $(ICON_WIN)

tools: $(TOOLS_DIR)/ginkgo $(TOOLS_DIR)/golangci-lint $(TOOLS_DIR)/goveralls $(TOOLS_DIR)/gover

$(TOOLS_DIR)/ginkgo:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/onsi/ginkgo/ginkgo@v1.12.0)

$(TOOLS_DIR)/golangci-lint:
	$(shell curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR) v1.25.0)

$(TOOLS_DIR)/goveralls:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/mattn/goveralls@v0.0.5)

$(TOOLS_DIR)/gover:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/modocache/gover)

$(TOOLS_DIR)/2goarray:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/cratonica/2goarray)
