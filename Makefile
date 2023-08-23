
server:
	go run main.go

test:
	go test -v -cover -short ./...

mock:
	mockgen -package mockplatform -destination platform/mock/platform.go github.com/yangwenz/model-serving/platform Platform

.PHONY: server test
