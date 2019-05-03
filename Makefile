all: build

GOPATH ?= $(shell mktemp -d)
PROGRAM ?= sk8
include hack/go.mk

.PHONY: $(PROGRAM)
$(PROGRAM):
	$(MAKE) fmt
	$(MAKE) -C ./pkg/config
	$(MAKE) -C ./pkg/provider/vsphere/config
	CGO_ENABLED=$(CGO_ENABLED) go build -mod vendor -ldflags '$(LDFLAGS)' -o $@

build: $(PROGRAM)

.PHONY: clean
clean:
	go clean -i -x

.PHONY: imports
imports:
	command -v goimports >/dev/null || { cd "$(GOPATH)" && GOOS=$(GOHOSTOS) GOARCH=$(GOHOSTARCH) go get golang.org/x/tools/cmd/goimports; }
	goimports -w ./pkg/cluster ./pkg/cmd ./pkg/config ./pkg/status ./pkg/util ./pkg/vsphere

.PHONY: fmt format
fmt format: gen
	command -v goformat >/dev/null || { cd "$(GOPATH)" && GOOS=$(GOHOSTOS) GOARCH=$(GOHOSTARCH) go get github.com/mbenkmann/goformat/goformat; }
	goformat -s -w ./pkg

.PHONY: gen generated
gen generated:
	$(MAKE) -C pkg/config

upload: build
	aws s3 cp $(PROGRAM) s3://cnx.vmware/$(PROGRAM) \
	  --grants read=uri=http://acs.amazonaws.com/groups/global/AllUsers && \
	  echo https://s3-us-west-2.amazonaws.com/cnx.vmware/$(PROGRAM)
.PHONY: upload
