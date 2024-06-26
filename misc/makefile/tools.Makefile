# This makefile should be used to hold functions/variables

ifeq ($(ARCH),x86_64)
	ARCH := amd64
else ifeq ($(ARCH),aarch64)
	ARCH := arm64 
endif

define github_url
    https://github.com/$(GITHUB)/releases/download/v$(VERSION)/$(ARCHIVE)
endef

# creates a directory bin.
bin:
	@ mkdir -p $@

# ~~~ Tools ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

# ~~ [ gotestsum ] ~~~ https://github.com/gotestyourself/gotestsum ~~~~~~~~~~~~~~~~~~~~~~~

GOTESTSUM := $(shell command -v gotestsum || echo "bin/gotestsum")
gotestsum: bin/gotestsum ## Installs gotestsum (testing go code)

bin/gotestsum: VERSION := 1.12.0
bin/gotestsum: GITHUB  := gotestyourself/gotestsum
bin/gotestsum: ARCHIVE := gotestsum_$(VERSION)_$(OSTYPE)_$(ARCH).tar.gz
bin/gotestsum: bin
	@ printf "Install gotestsum... "
	@ printf "$(github_url)\n"
	@ curl -Ls $(shell echo $(call github_url) | tr A-Z a-z) | tar -zOxf - gotestsum > $@ && chmod +x $@
	@ echo "done."

# ~~ [ tparse ] ~~~ https://github.com/mfridman/tparse ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

TPARSE := $(shell command -v tparse || echo "bin/tparse")
tparse: bin/tparse ## Installs tparse (testing go code)

# eg https://github.com/mfridman/tparse/releases/download/v0.13.2/tparse_darwin_arm64
export TPARSE_ARCH := $(shell uname -m)
bin/tparse: VERSION := 0.13.3
bin/tparse: GITHUB  := mfridman/tparse
bin/tparse: ARCHIVE := tparse_$(OSTYPE)_$(TPARSE_ARCH) #this is custom
bin/tparse: bin
	@ printf "Install tparse... "
	@ printf "$(github_url)\n"
	@ curl -Ls $(call github_url) > $@ && chmod +x $@
	@ echo "done."

# ~~ [ mockery ] ~~~ https://github.com/vektra/mockery ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

MOCKERY := $(shell command -v mockery || echo "bin/mockery")
mockery: bin/mockery ## Installs mockery (mocks generation)

export MOCKERY_ARCH := $(shell uname -m)
export MOCKERY_OSTYPE := $(shell uname -s)
bin/mockery: VERSION := 2.43.2
bin/mockery: GITHUB  := vektra/mockery
bin/mockery: ARCHIVE := mockery_$(VERSION)_$(MOCKERY_OSTYPE)_$(MOCKERY_ARCH).tar.gz
bin/mockery: bin
	@ printf "Install mockery... "
	@ printf "$(github_url)\n"
	@ curl -Ls $(call github_url) | tar -zOxf -  mockery > $@ && chmod +x $@
	@ echo "done."

# ~~ [ golangci-lint ] ~~~ https://github.com/golangci/golangci-lint ~~~~~~~~~~~~~~~~~~~~~

GOLANGCI := $(shell command -v golangci-lint || echo "bin/golangci-lint")
golangci-lint: bin/golangci-lint ## Installs golangci-lint (linter)

bin/golangci-lint: VERSION := 1.59.0
bin/golangci-lint: GITHUB  := golangci/golangci-lint
bin/golangci-lint: ARCHIVE := golangci-lint-$(VERSION)-$(OSTYPE)-$(ARCH).tar.gz
bin/golangci-lint: bin
	@ printf "Install golangci-linter... "
	@ printf "$(github_url)\n"
	@ curl -Ls $(shell echo $(call github_url) | tr A-Z a-z) | tar -zOxf - $(shell printf golangci-lint-$(VERSION)-$(OSTYPE)-$(ARCH)/golangci-lint | tr A-Z a-z ) > $@ && chmod +x $@
	@ echo "done."