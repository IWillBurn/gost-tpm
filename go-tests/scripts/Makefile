BINARY_NAME = go-tests
GOOS = linux
GOARCH = amd64
BUILD_DIR = build

all: build

build-tests:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -c -o $(BUILD_DIR)/$(BINARY_NAME) ./tests/constants.go ./tests/fake_hashing_test.go ./tests/has_algs_test.go

.PHONY: all build