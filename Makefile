.PHONY: check
lint:
	golangci-lint run

test:
	go test -mod=mod -count=1 --race ./...

test-ci:
	go test -mod=mod -count=1 -timeout 10s  -coverprofile=coverage.txt -covermode=atomic ./...
