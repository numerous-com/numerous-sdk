SHELL = bash
TARGET_SYSTEMS := darwin windows linux
TARGET_ARCHS := amd64 arm64

# CLI related variables
GO_ENV = CGO_ENABLED=0
GO_BUILD = cd cli && $(GO_ENV) go build

CLI_DIR=./cli
CLI_BUILD_DIR_NAME=build
CLI_BUILD_DIR=$(CLI_DIR)/$(CLI_BUILD_DIR_NAME)
CLI_SOURCE_FILES=$(shell find $(CLI_DIR) -name *.go -type f)

CLI_BUILD_TARGETS := $(foreach SYS,$(TARGET_SYSTEMS),$(foreach ARCH,$(TARGET_ARCHS),$(CLI_BUILD_DIR)/$(SYS)_$(ARCH)))

get_cli_target_from_sdk_binary = $(word 1,$(subst $(SDK_CLI_BINARY_DIR)/,,$(CLI_BUILD_DIR)/$@))
getsystem = $(word 3,$(subst _, ,$(subst /, ,$@)))
getarch = $(word 4,$(subst _, ,$(subst /, ,$@)))

GQL_HTTP_URL = https://api.numerous.com/query
GQL_WS_URL = wss://api.numerous.com/query
AUTH0_DOMAIN = numerous.eu.auth0.com
AUTH0_CLIENT_ID = h5U41HhtgJ5OXdIvzi2Aw7VNFQMoLzgF
LDFLAGS = -s -w \
          -X "numerous/cli/internal/gql.httpURL=$(GQL_HTTP_URL)" \
          -X "numerous/cli/internal/gql.wsURL=$(GQL_WS_URL)" \
		  -X "numerous/auth.auth0Domain=$(AUTH0_DOMAIN)" \
		  -X "numerous/auth.auth0ClientID=$(AUTH0_CLIENT_ID)"


# Python SDK related variables
SDK_CLI_BINARY_DIR=python/src/numerous/cli/build
SDK_CLI_BINARY_TARGETS := $(foreach SYS,$(TARGET_SYSTEMS),$(foreach ARCH,$(TARGET_ARCHS),$(SDK_CLI_BINARY_DIR)/$(SYS)_$(ARCH)))
SDK_CHECK_VENV=@if [ -z "${VIRTUAL_ENV}" ]; then echo "-- Error: An activated virtual environment is required"; exit 1; fi

# RULES
.DEFAULT_GOAL := help

.PHONY: clean test lint dep package sdk-binaries sdk-test sdk-lint sdk-dep cli-test cli-lint cli-dep cli-all cli-build cli-local

clean:
	rm -rf $(CLI_BUILD_DIR)
	rm -rf $(SDK_CLI_BINARY_DIR)
	rm -rf dist
	rm -f .lint-ruff.txt
	rm -f .lint-mypy.txt

package: sdk-binaries
	@echo "-- Building SDK package"
	python -m build

test: sdk-test cli-test

lint: sdk-lint cli-lint

dep: sdk-dep cli-dep

sdk-lint:
	@echo "-- Running SDK linters"
	$(SDK_CHECK_VENV)
	ruff check . && echo -n "true" > .lint-ruff.txt || echo -n "false" > .lint-ruff.txt;
	mypy --strict . && echo -n "true" > .lint-mypy.txt || echo -n "false" > .lint-mypy.txt;
	$$(cat .lint-ruff.txt) && $$(cat .lint-mypy.txt)

sdk-test:
	@echo "-- Running tests for sdk"
	$(SDK_CHECK_VENV)
	pytest python/tests

sdk-dep:
	@echo "-- Installing SDK dependencies"
	$(SDK_CHECK_VENV)
	pip install -e .[dev] -q

sdk-binaries: $(SDK_CLI_BINARY_TARGETS)

# Directory for CLI binaries, in SDK
$(SDK_CLI_BINARY_DIR):
	@echo "-- Creating SDK binary directory"
	mkdir $(SDK_CLI_BINARY_DIR)

# CLI executables in SDK
$(SDK_CLI_BINARY_DIR)/%: $(SDK_CLI_BINARY_DIR) $(CLI_BUILD_DIR)/%
	echo "Copying built binary $@"
	cp $(get_cli_target_from_sdk_binary) $@

# CLI for specific OS/architecture
cli-all: $(CLI_BUILD_TARGETS)

$(CLI_BUILD_TARGETS): %: $(CLI_SOURCE_FILES)
	@echo "-- Building CLI for OS $(getsystem) architecture $(getarch) in $@"
	export GOARCH=$(getarch) GOOS=$(getsystem) && $(GO_BUILD) -ldflags '$(LDFLAGS)' -o ../$@ .

cli-local:
	@echo "-- Building local CLI"
	$(GO_BUILD) -o $(CLI_BUILD_DIR_NAME)/local .

cli-build:
	@echo "-- Building CLI"
	$(GO_BUILD) -ldflags '$(LDFLAGS)' -o $(CLI_BUILD_DIR_NAME)/numerous .

cli-lint:
	@echo "-- Running CLI linters"
	cd cli && golangci-lint run
	cd cli && gofumpt -l -w .

cli-test:
	@echo "-- Running CLI tests"
	cd cli && go test -coverprofile=c.out ./...

cli-dep:
	@echo "-- Installing CLI dependencies"
	cd cli && go mod download
	cd cli && go mod tidy  > /dev/null

gqlgen:
	@echo "-- Generating GraphQL code"
	cd python && ariadne-codegen

help:
	@echo "Make targets (help is default):"
	@echo "    test         Run all tests"
	@echo "    lint         Run all linters"
	@echo "    dep          Install all dependencies"
	@echo "    package      Package the SDK python package including CLI builds"
	@echo "    sdk-binaries Build CLI binaries in SDK package"
	@echo "    sdk-test     Run SDK tests"
	@echo "    sdk-lint     Run SDK linters"
	@echo "    sdk-dep      Install SDK dependencies"
	@echo "    cli-test     Run CLI tests"
	@echo "    cli-lint     Run CLI linters"
	@echo "    cli-dep      Install CLI dependencies"
	@echo "    cli-all      Build CLI for all systems"
	@echo "    cli-build    Build CLI for current system"
	@echo "    cli-local    Build local CLI for current system"
	@echo "    help         Display this message"
