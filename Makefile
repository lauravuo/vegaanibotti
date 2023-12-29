fumpt:
	gofumpt -l -w .

lint:
	golangci-lint run

test:
	go test ./...

fetch:
	go run . --fetch
