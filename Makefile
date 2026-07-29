.PHONY: all build test lint vet vuln clean run-mock help check

# ── variables ────────────────────────────────────────────────
BINARY     := barrage
CMD        := ./cmd/barrage
MOCK_CMD   := ./cmd/mockserver
GO         := go
GOFLAGS    :=

# ── default ──────────────────────────────────────────────────
all: vet lint test build

# ── build ────────────────────────────────────────────────────
build:
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD)

build-mock:
	$(GO) build $(GOFLAGS) -o mockserver $(MOCK_CMD)

# ── test ─────────────────────────────────────────────────────
test:
	$(GO) test ./... -race -count=1

test-verbose:
	$(GO) test ./... -race -count=1 -v

test-cover:
	$(GO) test ./... -race -count=1 -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "coverage report: coverage.html"

# ── quality ──────────────────────────────────────────────────
vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

# ── clean ────────────────────────────────────────────────────
clean:
	rm -f $(BINARY) mockserver coverage.out coverage.html

# ── dev helpers ──────────────────────────────────────────────
run-mock:
	$(GO) run $(MOCK_CMD) -port 8090 -latency 5ms -error-rate 0.02

run-example:
	$(GO) run $(CMD) -url http://localhost:8090 -requests 500 -workers 20 -rate 100

# ── full check (CI equivalent) ────────────────────────────────
check: vet lint vuln test
	@echo "✓ all checks passed"

# ── help ─────────────────────────────────────────────────────
help:
	@echo ""
	@echo "  make build        build barrage binary"
	@echo "  make build-mock   build mockserver binary"
	@echo "  make test         run tests with -race"
	@echo "  make test-cover   run tests + open coverage report"
	@echo "  make vet          go vet"
	@echo "  make lint         golangci-lint"
	@echo "  make vuln         govulncheck"
	@echo "  make all          vet + lint + test + build"
	@echo "  make run-mock     start mock server"
	@echo "  make run-example  run barrage against mock server"
	@echo "  make clean        remove build artifacts"
	@echo ""