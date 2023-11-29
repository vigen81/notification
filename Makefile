GOPATH:=$(shell go env GOPATH)

.PHONY: init
init:
	@go get -u google.golang.org/protobuf/proto
	@go install github.com/golang/protobuf/protoc-gen-go@latest
	@go install github.com/asim/go-micro/cmd/protoc-gen-micro/v4@latest

.PHONY: proto
proto:
	@protoc \
	--go_opt=paths=source_relative \
	--micro_opt=paths=source_relative \
	--proto_path=./proto \
	--proto_path=/Users/vahan/Development/googleapis \
	--experimental_allow_proto3_optional \
	--micro_out=./pb \
	--go_out=:./pb ./proto/$(type)/*.proto

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: build
build:
	@go build -o notification-service ./

.PHONY: test
test:
	@go test -v ./... -cover

.PHONY: docker
docker:
	@docker  build  --build-arg gitlab_user="boot" --build-arg gitlab_personal_token="fdVN9c6AGgz9c9WR9P7V" -t registry.gitlab.com/healthcare-integration/golang/notification-service:dev .

.PHONY: push
push:
	@docker push registry.gitlab.com/healthcare-integration/golang/notification-service:dev

.PHONY: models
models:
	@go generate ./ent
