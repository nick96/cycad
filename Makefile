PROTOC=bin/protoc

PROTOC_ARGS=-I./third_party/googleapis

EDITOR_PROTO=protos/editor_service.proto
HEALTH_PROTO=protos/health_service.proto

all: editor health gateway

editor-pb:
	$(PROTOC) $(PROTOC_ARGS) -Iprotos --go_out=plugins=grpc,paths=source_relative:./editorservice/pb $(EDITOR_PROTO)

editor-gateway-pb:
	$(PROTOC) $(PROTOC_ARGS) -Iprotos --grpc-gateway_out=logtostderr=true,paths=source_relative:./gateway/pb $(EDITOR_PROTO)
	$(PROTOC) $(PROTOC_ARGS) -Iprotos --go_out=plugins=grpc,paths=source_relative:./gateway/pb $(EDITOR_PROTO)

editor: editor-pb
	cd editorservice && go build github.com/nick96/cycad/editorservice

health-pb:
	$(PROTOC) $(PROTOC_ARGS) -Iprotos --go_out=plugins=grpc,paths=source_relative:./healthservice/pb $(HEALTH_PROTO)

health-gateway-pb:
	$(PROTOC) $(PROTOC_ARGS) -Iprotos --grpc-gateway_out=logtostderr=true,paths=source_relative:./gateway/pb $(HEALTH_PROTO)
	$(PROTOC) $(PROTOC_ARGS) -Iprotos --go_out=plugins=grpc,paths=source_relative:./gateway/pb $(HEALTH_PROTO)

gateway-pb: editor-gateway-pb health-gateway-pb

health: health-pb
	cd healthservice && go build github.com/nick96/cycad/healthservice

gateway: gateway-pb
	cd gateway && go build github.com/nick96/cycad/gateway

