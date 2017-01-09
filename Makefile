VERSION?=$(shell git describe --tags --always)
GO_FILES:=$(shell find . -not -path "./vendor/*" -name "*.go")
GO_PROGRAM_FILES:=$(shell find . -not -path "./vendor/*" -not -name "*_test.go" -name "*.go")
GOIMPORTS?=$(shell goimports -d $(GO_FILES) | tee /dev/stderr)
TMP_DIR:=.tmp
BUILD_DIR:=dist
BUILD_OS:="linux freebsd"
BUILD_ARCH:="amd64 386"
BIN_NAME:=awslogs

LD_FLAGS:="-X main.version=$(VERSION)"

PKG_gox=github.com/mitchellh/gox
PKG_goveralls=github.com/mattn/goveralls
PKG_goimports=golang.org/x/tools/cmd/goimports

.DEFAULT_GOAL := bin

coveralls: dep
	@goveralls -service=travis-ci

dep:
	@go get \
		 $(PKG_goimports) \
		 $(PKG_goveralls) \
		 $(PKG_gox)

bin:
	@go build \
		-race \
		-tags debug \
		-o $(TMP_DIR)/$(BIN_NAME) \
		-ldflags $(LD_FLAGS)

run: bin
	@$(TMP_DIR)/$(BIN_NAME)

test: dep
	@if [ "$(GOIMPORTS)" ]; then exit 1; fi
	@go vet
	@go test -race

dist: dep distclean
	@gox \
		-os=$(BUILD_OS) \
		-arch=$(BUILD_ARCH) \
		-output "$(BUILD_DIR)/{{.OS}}_{{.Arch}}_$(BIN_NAME)" \
		-ldflags $(LD_FLAGS)

distclean:
	@rm -f $(BUILD_DIR)/*
