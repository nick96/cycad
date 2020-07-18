PROTOC_ARGS=-Iprotos -Ithird_party/googleapis

EDITOR_PROTO=protos/services/editor/v1/editorservice.proto

editor-stub:
	./bin/protoc $(PROTOC_ARGS) --go_out=plugins=grpc,path=source_relative:. $(EDITOR_PROTO)

editor-gateway:
	./bin/protoc $(PROTOC_ARGS) --grpc-gateway_out=logtostderr=true,paths=source_relative:. $(EDITOR_PROTO)

editor-swagger:
	./bin/protoc $(PROTOC_ARGS) --swagger_out=logtostderr=true:./editorservice/ $(EDITOR_PROTO)
