all: client server
	@echo "All done for " $(GOOS) ":" $(GOARCH)

client: .PHONY
	go build -v -o client app_client/main.go

server: .PHONY
	go build -v -o server app_server/main.go

.PHONY:

test:
	go test
