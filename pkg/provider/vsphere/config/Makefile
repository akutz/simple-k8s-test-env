all: build

ROOT_DIR := $(abspath ../../../..)
include $(ROOT_DIR)/hack/go.mk
BOILERPLATE ?= $(ROOT_DIR)/hack/boilerplate/boilerplate.generated.go.txt
DEEPCOPY_BIN ?= $(ROOT_DIR)/hack/.build/$(GOHOSTOS)_$(GOHOSTARCH)/deepcopy-gen
DEFAULTER_BIN ?= $(ROOT_DIR)/hack/.build/$(GOHOSTOS)_$(GOHOSTARCH)/defaulter-gen
PACKAGE := $(shell go list -mod vendor .)
SRCS := $(filter-out zz_%.go,$(wildcard *.go))
ZZ_DEEPCOPY := zz_generated.deepcopy.go
ZZ_DEFAULT := zz_generated.default.go

build: $(ZZ_DEEPCOPY)
	CGO_ENABLED=$(CGO_ENABLED) go build -mod vendor -ldflags '$(LDFLAGS)' .

$(DEEPCOPY_BIN) $(DEFAULTER_BIN):
	GOOS=$(GOHOSTOS) GOARCH=$(GOHOSTARCH) go build -mod vendor -o $@ $(ROOT_DIR)/vendor/k8s.io/code-generator/cmd/$(@F)

$(ZZ_DEEPCOPY): $(SRCS) | $(DEEPCOPY_BIN)
	@echo "Generating DeepCopy() methods..." >&2 && \
	$(word 1,$|) \
	  --go-header-file $(BOILERPLATE) \
	  --input-dirs $(PACKAGE) \
	  --output-file-base $(subst .go,,$(@F)) \
	  --bounding-dirs $(PACKAGE)

ifneq (,$(wildcard default.go))
build: $(ZZ_DEFAULT)
$(ZZ_DEFAULT): $(SRCS) $(ZZ_DEEPCOPY) | $(DEFAULTER_BIN)
	@echo "Generating Default() methods..." >&2 && \
	$(word 1,$|) \
	  --go-header-file $(BOILERPLATE) \
	  --input-dirs $(PACKAGE) \
	  --output-file-base $(subst .go,,$(@F))
endif
