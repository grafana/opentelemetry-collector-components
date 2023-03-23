#include ../../Makefile.Common

.PHONY: proto
proto:
	protoc --go_out=lookup/ --go_opt=paths=source_relative --go-grpc_out=lookup/ --go-grpc_opt=paths=source_relative proto/lookup.proto
