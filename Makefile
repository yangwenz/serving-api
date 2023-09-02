
server:
	go run main.go

test:
	go test -v -cover -short ./...

mock:
	mockgen -package mockplatform -destination platform/mock/platform.go github.com/yangwenz/model-serving/platform Platform
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/yangwenz/model-serving/worker TaskDistributor

docker:
	docker build -t yangwenz/model-serving:latest .
	docker push yangwenz/model-serving:latest

.PHONY: server test mock docker
