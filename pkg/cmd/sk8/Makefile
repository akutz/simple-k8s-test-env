all: build

sk8:
	make -C ../../app resgen_sk8.go
	CGO_ENABLED=0 go build -mod vendor -tags netgo -ldflags "-w"

build: sk8

clean:
	go clean -i -x

.PHONY: sk8