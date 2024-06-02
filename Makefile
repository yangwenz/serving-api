
server:
	go run main.go

test:
	go test -v -cover -short ./...

mock:
	mockgen -package mockapi -destination api/mock/webhook.go github.com/HyperGAI/serving-api/api Webhook
	mockgen -package mockapi -destination api/mock/auth.go github.com/HyperGAI/serving-api/api Authenticator

docker:
	docker build --platform=linux/amd64 -t yangwenz/serving-api:v1 .
	docker push yangwenz/serving-api:v1

.PHONY: server test docker mock
