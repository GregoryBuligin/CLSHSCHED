BIN_NAME = clshshed
BIN_CLIENT_NAME = clctl

.PHONY: run
run:
	go run main.go

.PHONY: build
build:
	go build -o ${BIN_NAME} main.go

.PHONY: build_client
build_client:
	go build -o ${BIN_CLIENT_NAME} client/main.go

.PHONY: protoc
protoc:
	protoc -I protobuf/ protobuf/shsched.proto --go_out=plugins=grpc:shsched

.PHONY: clean
clean:
	-rm ${BIN_NAME}
	-rm ${BIN_CLIENT_NAME}
