SHELL := /bin/bash
.SHELLFLAGS := -euo pipefail -c

GO_PACKAGES := ./...
COVERAGE_PROFILE := coverage.out
BENCH_OUTPUT := benchmarks/current.txt

.PHONY: test test-race lint coverage bench ci

test:
	go test $(GO_PACKAGES)

test-race:
	go test -race -count=1 $(GO_PACKAGES)

lint:
	golangci-lint run $(GO_PACKAGES)

coverage:
	go test -coverprofile=$(COVERAGE_PROFILE) -cover $(GO_PACKAGES)
	go-test-coverage --config=./.testcoverage.yml

bench:
	mkdir -p $(dir $(BENCH_OUTPUT))
	go test -bench=. -benchmem -count=1 -run='^$$' $(GO_PACKAGES) | tee $(BENCH_OUTPUT)

ci:
	$(MAKE) test
	$(MAKE) test-race
	$(MAKE) lint
	$(MAKE) coverage
