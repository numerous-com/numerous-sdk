SHELL = bash
TARGET_SYSTEMS := darwin windows linux
TARGET_ARCHS := amd64 arm64
TARGETS := $(foreach SYS,$(TARGET_SYSTEMS),$(foreach ARCH,$(TARGET_ARCHS),$(SYS)_$(ARCH)))

# CLI related variables
GO_ENV = CGO_ENABLED=0
GO_BUILD = $(GO_ENV) go build
CLI_BUILD_DIR=build
CLI_SOURCE_FILES=$(shell find . -name '*.go' -type f)
CLI_BUILD_TARGETS := $(foreach TARGET,$(TARGETS),$(CLI_BUILD_DIR)/$(TARGET))
VERSION_TXT=internal/version/version.txt

get_cli_target_from_sdk_binary = $(word 1,$(subst $(SDK_CLI_BINARY_DIR)/,,$(CLI_BUILD_DIR)/$@))
getsystem = $(word 2,$(subst _, ,$(subst /, ,$@)))
getarch = $(word 3,$(subst _, ,$(subst /, ,$@)))

GQL_HTTP_URL = https://api.numerous.com/query
GQL_WS_URL = wss://api.numerous.com/query
AUTH0_DOMAIN = auth.numerous.com
AUTH0_CLIENT_ID = h5U41HhtgJ5OXdIvzi2Aw7VNFQMoLzgF
AUTH0_AUDIENCE = https://numerous.eu.auth0.com/api/v2/
LDFLAGS = -s -w \
          -X "numerous.com/cli/internal/gql.httpURL=$(GQL_HTTP_URL)" \
          -X "numerous.com/cli/internal/gql.wsURL=$(GQL_WS_URL)" \
		  -X "numerous.com/cli/internal/auth.auth0Domain=$(AUTH0_DOMAIN)" \
		  -X "numerous.com/cli/internal/auth.auth0ClientID=$(AUTH0_CLIENT_ID)" \
		  -X "numerous.com/cli/internal/auth.auth0Audience=$(AUTH0_AUDIENCE)"


# Python SDK related variables
SDK_CLI_BINARY_DIR=python/src/numerous/cli/bin
SDK_CHECK_VENV=@if [ -z "${VIRTUAL_ENV}" ]; then echo "-- Error: An activated virtual environment is required"; exit 1; fi

# Packaging
PACKAGE_TARGETS := $(foreach TARGET,$(TARGETS),package-$(TARGET))

# Version
create_version_txt_cmd=grep '^version = ".\+"' pyproject.toml | tr -d '\n' | sed 's/^version = "\(.\+\)"/\1/' > $(VERSION_TXT)

# RULES
.DEFAULT_GOAL := help

.PHONY: clean packages test lint dep sdk-test sdk-lint sdk-dep cli-test cli-lint cli-dep cli-all cli-build cli-local version $(PACKAGE_TARGETS)

clean:
	rm -rf $(CLI_BUILD_DIR)
	rm -rf $(SDK_CLI_BINARY_DIR)
	rm -rf dist
	rm -f .lint-ruff.txt
	rm -f .lint-mypy.txt
	rm -f $(VERSION_TXT)

packages: ${PACKAGE_TARGETS}

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

# Directory for CLI binaries, in SDK
$(SDK_CLI_BINARY_DIR):
	@echo "-- Creating SDK binary directory"
	mkdir $(SDK_CLI_BINARY_DIR)

# CLI executables in SDK
$(SDK_CLI_BINARY_DIR)/%: $(SDK_CLI_BINARY_DIR) $(CLI_BUILD_DIR)/%
	echo "-- Copying built binary $@"
	cp $(get_cli_target_from_sdk_binary) $@

$(PACKAGE_TARGETS): package-% : $(CLI_BUILD_DIR)/%
	echo "-- Building package $*"
	scripts/build_dists.sh $(notdir $*)

package-any: $(CLI_BUILD_TARGETS)
	echo "-- Building 'any' package"
	scripts/build_dists.sh any

# CLI for specific OS/architecture
cli-all: $(CLI_BUILD_TARGETS)

$(CLI_BUILD_TARGETS): %: $(CLI_SOURCE_FILES)
	@echo "-- Building CLI for OS $(getsystem) architecture $(getarch) in $@"
	$(create_version_txt_cmd)
	export GOARCH=$(getarch) GOOS=$(getsystem) && $(GO_BUILD) -ldflags '$(LDFLAGS)' -o $@ .

cli-local:
	@echo "-- Building local CLI"
	$(create_version_txt_cmd)
	$(GO_BUILD) -o $(CLI_BUILD_DIR)/local .

cli-build:
	@echo "-- Building CLI"
	$(create_version_txt_cmd)
	$(GO_BUILD) -ldflags '$(LDFLAGS)' -o $(CLI_BUILD_DIR)/numerous .

cli-lint:
	@echo "-- Running CLI linters"
	$(create_version_txt_cmd)
	golangci-lint run
	gofumpt -l -w .

cli-test:
	@echo "-- Running CLI tests"
	$(create_version_txt_cmd)
	gotestsum -f testname -- -coverprofile=c.out ./...

cli-dep:
	@echo "-- Installing CLI dependencies"
	go mod download
	go mod tidy  > /dev/null

gqlgen:
	@echo "-- Generating GraphQL code"
	cd python && ariadne-codegen

version:
	$(create_version_txt_cmd)

help:
	@echo "Make targets (help is default):"
	@echo "    clean        Clean all build artifacts"
	@echo "    test         Run all tests"
	@echo "    lint         Run all linters"
	@echo "    dep          Install all dependencies"
	@echo "    packages     Create all SDK distributions with CLI binaries"
	@echo "    pkg-PLAT     Create SDK distribution for platform PLAT, e.g. linux_amd64"
	@echo "    sdk-test     Run SDK tests"
	@echo "    sdk-lint     Run SDK linters"
	@echo "    sdk-dep      Install SDK dependencies"
	@echo "    cli-test     Run CLI tests"
	@echo "    cli-lint     Run CLI linters"
	@echo "    cli-dep      Install CLI dependencies"
	@echo "    cli-all      Build CLI for all systems"
	@echo "    cli-build    Build CLI for current system"
	@echo "    cli-local    Build local CLI for current system"
	@echo "    gqlgen       Generate graphql code"
	@echo "    version      Generate version file for embedding in CLI"
	@echo "    help         Display this message"
