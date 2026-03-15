# wave-core justfile
# https://github.com/casey/just

set dotenv-load := false

# project metadata
mod     := "github.com/wave-cli/wave-core"
bin     := "wave"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

# ─── default ──────────────────────────────────────────────────────────

# list available recipes
default:
    @just --list

# ─── build & run ──────────────────────────────────────────────────────

# build the wave binary
build:
    go build -ldflags "-X {{mod}}/internal/version.Version={{version}}" -o {{bin}} .

# run wave with arguments (e.g. just run version)
run *args:
    go run -ldflags "-X {{mod}}/internal/version.Version={{version}}" . {{args}}

# build and install to $GOPATH/bin
install:
    go install -ldflags "-X {{mod}}/internal/version.Version={{version}}" .

# remove build artifacts
clean:
    rm -f {{bin}}
    rm -f coverage.out coverage.html
    rm -f testdata/plugins/echo/echo

# ─── test ─────────────────────────────────────────────────────────────

# run all tests
test:
    go test ./...

# run tests with verbose output
test-v:
    go test -v ./...

# run tests with coverage report
test-cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# open coverage report in browser
test-cover-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    xdg-open coverage.html 2>/dev/null || open coverage.html 2>/dev/null || echo "Open coverage.html in your browser"

# run only unit tests (skip e2e)
test-unit:
    go test $(go list ./... | grep -v /e2e)

# run only e2e tests
test-e2e:
    go test -v ./e2e/...

# run tests for a specific package (e.g. just test-pkg config)
test-pkg pkg:
    go test -v ./internal/{{pkg}}/...

# ─── lint & format ───────────────────────────────────────────────────

# format all go files
fmt:
    go fmt ./...

# vet all packages
vet:
    go vet ./...

# run fmt + vet
lint: fmt vet

# ─── docs ─────────────────────────────────────────────────────────────

# open architecture doc
docs:
    xdg-open docs/architecture.md 2>/dev/null || open docs/architecture.md 2>/dev/null || less docs/architecture.md

# open testing guide
docs-testing:
    xdg-open docs/testing.md 2>/dev/null || open docs/testing.md 2>/dev/null || less docs/testing.md

# open plugin authoring guide
docs-plugin:
    xdg-open docs/plugin-authoring.md 2>/dev/null || open docs/plugin-authoring.md 2>/dev/null || less docs/plugin-authoring.md

# open the project plan
docs-plan:
    xdg-open plan.md 2>/dev/null || open plan.md 2>/dev/null || less plan.md

# ─── deps ─────────────────────────────────────────────────────────────

# tidy go modules
tidy:
    go mod tidy

# download dependencies
deps:
    go mod download

# ─── dev helpers ──────────────────────────────────────────────────────

# build the echo test plugin
build-echo:
    go build -o testdata/plugins/echo/echo ./testdata/plugins/echo/

# run the full ci pipeline locally (fmt, vet, test, build)
ci: lint test build
    @echo "✓ CI passed"
