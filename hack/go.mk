export CGO_ENABLED ?= 0
export LDFLAGS ?= -extldflags "-static" -w -s
export GOHOSTOS ?= $(shell go env GOHOSTOS)
export GOHOSTARCH ?= $(shell go env GOHOSTARCH)