.PHONY: build

build:
	go build ./cmd/...

test:
	go test `glide novendor`
