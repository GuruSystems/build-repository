PROTOCINC = -I.
PROTOCINC += -I${GOPATH}/src/
PROTOCINC += -I${GOPATH}/src/github.com/googleapis/googleapis/third_party/protobuf/src/
PROTOCINC += -I${GOPATH}/src/github.com/googleapis/googleapis/
PROTOCINC += -I${GOPATH}/src/github.com/googleapis/googleapis/third_party/
PROTOCINC += -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis

server:
	go install build-repo-server.go 
client:
	go install build-repo-client.go
frontend:
	go install build-repo-fe.go

.PHONY: proto
proto:
	@echo compiling Go proto stubs
	@protoc ${PROTOCINC} --go_out=plugins=grpc:. proto/build-repo.proto
	# for json gateway to compile you need the plugins:
	# 1. protoc-gen-swagger
	# 2. protoc-gen-grpc-gateway
	# they are in package github.com/grpc-ecosystem/grpc-gateway
#	@echo compiling json api gateway stuff
#	@protoc ${PROTOCINC} --grpc-gateway_out=logtostderr=true:. proto/build-repo.proto
#	@echo compiling hipster swagger stuff
#	@protoc ${PROTOCINC} --swagger_out=logtostderr=true:. proto/build-repo.proto
	@echo see files in directory proto/
	# java client: not enabled
	@#protoc ${PROTOCINC} --java_out=. proto/build-repo.proto
	# php client: not enabled
	@#https://grpc.io/docs/tutorials/basic/php.html
