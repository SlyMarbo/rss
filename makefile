all:
	make format
	make test

format:
	gofmt -s=true -w=true *.go

test:
	go test
