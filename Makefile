BIN := varnam-desktop
STUFFED_BIN := varnam
HASH := $(shell git rev-parse HEAD | cut -c 1-8)
BUILD_DATE := $(shell date '+%Y-%m-%d %H:%M:%S')
ARCH := $(shell uname -m)
COMMIT_DATE := $(shell git show -s --format=%ci ${HASH})
VERSION := ${HASH} (${COMMIT_DATE})
PRETTY_VERSION := $(shell git describe --abbrev=0 --tags)
RELEASE_NAME := varnam-${PRETTY_VERSION}
STATIC := ui:/
LDFLAGS := -X 'main.buildVersion=${VERSION}' -X 'main.buildDate=${BUILD_DATE}' -s -w

ifeq ($(OS),Windows_NT)
OS := windows
LDFLAGS := $(LDFLAGS) -H windowsgui
BIN := $(BIN).exe
RELEASE_NAME_32 := $(RELEASE_NAME)-windows-32
RELEASE_NAME := $(RELEASE_NAME)-windows-${ARCH}
else
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
OS := linux
RELEASE_NAME := $(RELEASE_NAME)-linux-${ARCH}
endif
ifeq ($(UNAME_S),Darwin)
OS := mac
RELEASE_NAME := $(RELEASE_NAME)-mac-${ARCH}
endif
endif

govarnam-linux:
	cd govarnam && make library && cp ./libgovarnam.so ../

govarnam-mac:
	cd govarnam && make library-mac-universal && cp ./libgovarnam.dylib ../

govarnam-windows:
	cd govarnam && make library && cp ./libgovarnam.so ../govarnam.dll

deps:
	go get -u github.com/knadh/stuffbin/...

.PHONY: build
build: ## Build the binary (default)
	go build -ldflags="${LDFLAGS}" -o ${BIN}
	stuffbin -a stuff -in ${BIN} -out ${STUFFED_BIN} ${STATIC}

# 32-bit releases are only for Windows
build-32:
	set GOARCH=386
	$(MAKE) build

release-linux:
	mkdir -p ${RELEASE_NAME}
	cp -r schemes ${RELEASE_NAME}
	cp varnam.sh varnam libgovarnam.* config.toml ${RELEASE_NAME}
	tar -cvzf ${RELEASE_NAME}.tar.gz ${RELEASE_NAME}

release-mac:
	mkdir -p ${RELEASE_NAME}
	cp -r schemes ${RELEASE_NAME}
	cp varnam libgovarnam.* config.toml ${RELEASE_NAME}
	cp varnam.sh ${RELEASE_NAME}/varnam.command
	tar -cvzf ${RELEASE_NAME}.tar.gz ${RELEASE_NAME}

release-windows:
	mkdir -p ${RELEASE_NAME}
	cp -r schemes ${RELEASE_NAME}
	cp -a varnam.exe windows-setup.bat libgovarnam.dll config.toml ${RELEASE_NAME}
	powershell "Compress-Archive -Force ${RELEASE_NAME} ${RELEASE_NAME}.zip"

release-windows-32:
	mkdir -p ${RELEASE_NAME_32}
	cp -r schemes ${RELEASE_NAME}
	cp -a varnam.exe windows-setup.bat libgovarnam.dll config.toml ${RELEASE_NAME_32}
	tar -acvf ${RELEASE_NAME_32}.zip ${RELEASE_NAME_32}

release-32:
	$(MAKE) govarnam-windows-32
	$(MAKE) build-32
	$(MAKE) release-windows-32

ifeq ($(OS),windows)
.PHONY: govarnam
govarnam: govarnam-windows

release-os:
	$(MAKE) release-windows

ui:
	build-editor.bat
endif

ifeq ($(OS),linux)
.PHONY: govarnam
govarnam: govarnam-linux

release-os:
	$(MAKE) release-linux

ui:
	./build-editor.sh
endif

ifeq ($(OS),mac)
.PHONY: govarnam
govarnam: govarnam-mac

release-os:
	$(MAKE) release-mac

ui:
	./build-editor.sh
endif

.PHONY: editor
editor:
	$(MAKE) ui

release:
	$(MAKE) govarnam
	$(MAKE) ui
	$(MAKE) build
	$(MAKE) release-os

.PHONY: run
run: build
	./${BIN}

.PHONY: clean
clean: ## Remove temporary files and the binary
	go clean
	git clean -fdx
	cd govarnam && git clean -fdx

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build