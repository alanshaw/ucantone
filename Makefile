.PHONY: cover

cover:
	mkdir -p coverage
	go test -coverprofile=./coverage/c.out -v ./...
	go tool cover -html=./coverage/c.out
