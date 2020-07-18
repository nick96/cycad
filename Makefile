PROTOC=bin/protoc

PROTOC_ARGS=-I./third_party/googleapis

EDITOR_PROTO=editorservice/service.proto

editor-stub:
	$(PROTOC) $(PROTOC_ARGS) -Ieditorservice --go_out=plugins=grpc,paths=source_relative:./editorservice/pb $(EDITOR_PROTO)

editor-gateway:
	$(PROTOC) $(PROTOC_ARGS) -Ieditorservice --grpc-gateway_out=logtostderr=true,paths=source_relative:./editorservice/pb $(EDITOR_PROTO)

editor-swagger:
	$(PROTOC) $(PROTOC_ARGS) -Ieditorservice --swagger_out=logtostderr=true:./editorservice/swagger $(EDITOR_PROTO)

editor: editor-stub editor-gateway editor-swagger
