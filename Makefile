.PHONY: build test cover gen codegen-build gen-check ci

build:
	go build ./...

test:
	go test ./...

cover:
	mkdir -p coverage
	go test -coverprofile=./coverage/c.out -v ./...
	go tool cover -html=./coverage/c.out

gen:
	go generate ./...

# Compile the code generators under the `codegen` build tag they actually run
# with. A normal `go build ./...` never sets this tag, so a generator that no
# longer compiles (e.g. it references a renamed or removed type) would otherwise
# go unnoticed until someone runs `go generate`.
codegen-build:
	go build -tags codegen $(shell go list ./... | grep '/gen$$')

# Regenerate everything and fail if the committed output changed, catching both
# generators that no longer compile and generated files that are out of date.
gen-check: gen
	@git diff --exit-code -- '*_gen.go' '*_gen.*.go' || \
		{ echo "generated files are out of date; run 'make gen' and commit the result" >&2; exit 1; }

# Aggregate target suitable for CI.
ci: build codegen-build gen-check test
