.PHONY: test demo fmt

test:
	go test ./...

fmt:
	gofmt -w .

demo:
	go run ./examples/demo
