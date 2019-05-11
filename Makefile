

.PHONY: run
run:
	go run main.go

.PHONY: protoc
protoc:
	protoc -I protobuf/ protobuf/shsched.proto --go_out=plugins=grpc:shsched/server
