.PHONY: check
lint:
	golangci-lint run

test:
	go test -mod=mod -count=1 --race ./...
