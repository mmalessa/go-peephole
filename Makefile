MAINDIR := ./
OUTPUT := peephole

.PHONY: test
test:
	test -z "$(shell gofmt -l *.go ./)"
	go test -v $(MAINDIR)

.PHONY: format
format:
	gofmt -w *.go $(MAINDIR)

.PHONY: build
build:
	go build -o $(OUTPUT) $(MAINDIR)