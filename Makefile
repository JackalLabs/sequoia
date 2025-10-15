export GO111MODULE = on

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

###############################################################################
###                                   All                                   ###
###############################################################################

ldflags = -X github.com/JackalLabs/sequoia/config.COMMIT=$(COMMIT) \
		  -X github.com/JackalLabs/sequoia/config.VERSION=$(VERSION)

all: lint test-unit

install:
	@go install -ldflags '$(ldflags)'

.PHONY: install


###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

tools:
	@go install github.com/client9/misspell/cmd/misspell
	@go install golang.org/x/tools/cmd/goimports

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

clean:
	rm -rf $(BUILDDIR)/

.PHONY: clean

###############################################################################
###                                Linting                                  ###
###############################################################################
golangci_lint_cmd=golangci-lint

format-tools:
	go install mvdan.cc/gofumpt@latest
	gofumpt -l -w .

lint: format-tools
	@echo "--> Running linter"
	$(golangci_lint_cmd) run --timeout=10m

lint-fix:
	@echo "--> Running linter"
	$(golangci_lint_cmd) run --fix --out-format=tab --issues-exit-code=0

.PHONY: lint lint-fix format-tools

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -path "./venv" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -path "./venv" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -path "./venv" | xargs goimports -w -local github.com/desmos-labs/cosmos-go-wallet
.PHONY: format

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test-unit:
	@echo "Executing unit tests..."
	@go test -mod=readonly -v -coverprofile coverage.txt ./...
.PHONY: test-unit
