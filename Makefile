MAIN_PACKAGES ?= $(shell go list -f '{{ .ImportPath }}:{{ .Name }}' ./cmd/... | grep ":main" | grep -v "/examples/" | cut -d: -f1)
LOCAL_MODULE ?= github.com/pixie-sh/core-go

GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)
GOTAGS ?= netgo,osusergo
GOSUMDB ?= sum.golang.org

ifeq ($(GOOS),darwin)
    CGO_ENABLED ?= 1
else
    CGO_ENABLED ?= 0
endif

NPROC ?= $(shell nproc 2>/dev/null || sysctl -n hw.ncpu)

MOCKS ?= internal

SCOPE ?= local

AWS_REGION ?= eu-west-3
AWS_PROFILE ?= default

BIN_PATH ?= bin

.PHONY: docs

###################################
######         BUILD         ######
###################################

clean:
	rm -rf bin

build: clean ensure-deps
	mkdir bin
	@echo "=== Building application ==="
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} GOTAGS=${GOTAGS} go build ./...
	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags=$(LDFLAGS) -o bin/ $(MAIN_PACKAGES)

ensure-deps:
	@echo "=== Ensuring Deps ==="
	@GOSUMDB=$(GOSUMDB) go version
	@GOSUMDB=$(GOSUMDB) go mod tidy
	@GOSUMDB=$(GOSUMDB) go get ./...
	@GOSUMDB=$(GOSUMDB) go env GOPRIVATE
	@GOSUMDB=$(GOSUMDB) go mod tidy

###################################
######        TESTING        ######
###################################

mock/%:
	rm -R $*/tests/mocks/*; \
	mockery -all -keeptree -dir $* -outpkg mocks -output $*/tests/mocks/

mocks:
	@echo "=== Generating mocks ==="
	$(MAKE) -j`nproc` $(patsubst %, mock/%, $(MOCKS))

# no mocks available yet
check: # mocks
	@echo "=== Checking code ==="
	golangci-lint run --timeout 3m

lint-fix:
	golangci-lint run --fix --timeout 6m

local-test:
	@echo "=== Start local services ==="
	docker compose up -d
	$(MAKE) test; docker compose down

# no mocks available yet
test: ensure-deps# mocks
	@echo "=== Running tests ==="
	go clean -testcache
	SCOPE=${SCOPE} go test -timeout 30s -failfast -race -coverprofile=test.cover -tags="${TAGS}" -run=${RUN} ./...

	go tool cover -func=test.cover
	rm -f test.cover

fix-imports:
	@find . -type f -name "*.go" -exec goimports -w {} \;
	@find . -name '*.go' -print0 | xargs -0 goimports -w -local "$(LOCAL_MODULE)"

check-imports:
	@echo "Checking import grouping..."
	@for f in $$(find . -type f -name '*.go' -not -path './vendor/*'); do \
		diff=$$(goimports -d -local "$(LOCAL_MODULE)" $$f); \
		if [ -n "$$diff" ]; then \
			printf "❌ Imports misordered in %s\n%s\n" "$$f" "$$diff"; \
			exit 1; \
		fi; \
	done;
	@echo "✅ All imports are correctly grouped."
	@goimports -w pkg infra