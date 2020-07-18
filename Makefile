PROTOC=bin/protoc

PROTOC_ARGS=-I./third_party/googleapis

EDITOR_PROTO=editorservice/service.proto

all: editor

editor-pb:
	$(PROTOC) $(PROTOC_ARGS) -Ieditorservice --go_out=plugins=grpc,paths=source_relative:./editorservice/pb $(EDITOR_PROTO)

editor-gateway-pb:
	$(PROTOC) $(PROTOC_ARGS) -Ieditorservice --grpc-gateway_out=logtostderr=true,paths=source_relative:./editorservice/pb $(EDITOR_PROTO)

editor-swagger:
	$(PROTOC) $(PROTOC_ARGS) -Ieditorservice --swagger_out=logtostderr=true:./editorservice/swagger $(EDITOR_PROTO)

editor-gateway: editor-gateway-pb editor-pb
	cd editorservice && go build -o gateway/gateway github.com/nick96/cycad/editorservice/gateway

editor-service: editor-pb
	cd editorservice && go build github.com/nick96/cycad/editorservice

editor: editor-swagger editor-gateway editor-service
