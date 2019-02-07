all: build

res_sk8.go: ../../sk8.sh
	go generate

GENERATED_GO_SRCS := res_sk8.go
GO_SRCS := $(wildcard *.go) $(wildcard ../res/*.go) $(wildcard ../config/*.go)
TEST_SRCS := $(GO_SRCS) $(GENERATED_GO_SRCS)
SRCS := $(filter-out %_test.go, $(TEST_SRCS))

TEST_TARGET := app.test
$(TEST_TARGET): $(TEST_SRCS)
	go test -c -o $@

TARGET := $(shell go list -f '{{.Target}}')
$(TARGET): $(SRCS)
	go install

build: $(TARGET) $(TEST_TARGET)

clean:
	rm -f res_sk8.go
	go clean -i -x

.PHONY: clean
