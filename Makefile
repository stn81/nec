MAKEFILE=$(realpath $(lastword $(MAKEFILE_LIST)))
HOME_DIR=$(shell dirname $(MAKEFILE))

OS := $(shell uname -s)

#程序名称
APP  = nec

GO	 = $(GOROOT)/bin/go
LINT = golangci-lint
OUT_DIR = $(HOME_DIR)/outputs
BIN_DIR = $(OUT_DIR)/bin
BIN  = $(BIN_DIR)/$(APP)

nec_SRC  = $(HOME_DIR)/app/nec/main.go

#程序版本号
VER_MAJOR   = 1
VER_MINOR   = 0
VER_PATCH   = 0

REVISION	= $(shell git rev-parse HEAD >/dev/null 2>&1 && echo "revision_found")
BUILD_DATE	= $(shell date +%Y-%m-%dT%H:%M:%S)

ifeq ($(strip $(REVISION)),)
REVISION 	= unknown
LAST_AUTHOR = unknown
LAST_DATE 	= unknown
else
REVISION	= $(shell git rev-parse HEAD)
LAST_AUTHOR	= $(shell git --no-pager show -s --format='%ae' $(REVISION))
LAST_T      = $(shell git --no-pager show -s --format='%at' $(REVISION))
ifeq ($(OS), Darwin)
LAST_DATE   = $(shell date -r $(LAST_T) +%Y-%m-%dT%H:%M:%S)
else
LAST_DATE   = $(shell date -d @$(LAST_T) +%Y-%m-%dT%H:%M:%S)
endif
endif

ifneq ($(wildcard $(HOME_DIR)/vendor/github.com/stn81/kate/*),)
KATE_APP_PKG=github.com/stn81/nec/vendor/github.com/stn81/kate/app
else
KATE_APP_PKG=github.com/stn81/kate/app
endif

LDFLAGS  = -X $(KATE_APP_PKG).VersionMajor=$(VER_MAJOR)
LDFLAGS += -X $(KATE_APP_PKG).VersionMinor=$(VER_MINOR)
LDFLAGS += -X $(KATE_APP_PKG).VersionPatch=$(VER_PATCH)
LDFLAGS += -X $(KATE_APP_PKG).Revision=$(REVISION)
LDFLAGS += -X $(KATE_APP_PKG).LastAuthor=$(LAST_AUTHOR)
LDFLAGS += -X $(KATE_APP_PKG).LastDate=$(LAST_DATE)
LDFLAGS += -X $(KATE_APP_PKG).BuildDate=$(BUILD_DATE)

PROTO_PATH := proto
PROTOS := $(shell find $(PROTO_PATH) -name *.proto)
PROTO_GO_SRC := $(PROTOS:.proto=.pb.go)

all: 
	@echo "make: PWD=$(shell pwd)"
	@echo "make: GOROOT=$(GOROOT)"
	@echo "make: GOPATH=$(GOPATH)"
	@echo "make: CGO_CFLAGS=$(CGO_CFLAGS)"
	@echo "make: CGO_LDFLAGS=$(CGO_LDFLAGS)"
	@echo "make: LIBRARY_PATH=$(LIBRARY_PATH)"
	@echo "make: LD_LIBRARY_PATH=$(LD_LIBRARY_PATH)"
	#$(GO) build -mod vendor -ldflags "$(LDFLAGS)" -o $(BIN) $(nec_SRC)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) $(nec_SRC)

proto: $(PROTO_GO_SRC)

%.pb.go: %.proto
	protoc --go_out=plugins=grpc:$(PROTO_PATH) -I $(PROTO_PATH) $<

lint:
	$(LINT) run

clean:
	rm -rf $(OUT_DIR)
