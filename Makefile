project := pascal-s
build_dir := ./build
bin_dir := $(build_dir)/bin
executable := $(bin_dir)/$(project)

.PHONY: all run test clean help

all: $(executable)

help:
	@echo "Try one of the following make goals:"
	@echo " * all   - build project"
	@echo " * run   - start the REPL"
	@echo " * test  - run test suite"
	@echo " * clean - delete build files"

$(executable): $(shell find . -name '*.go' -not -path './build/*')
	@mkdir -p $(bin_dir)
	go build -o $(executable) ./cmd/pascal-s

run: $(executable)
	@$(executable)

test:
	go test ./...

clean:
	rm -rf $(build_dir)
