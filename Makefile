#

PROTO = protobuf/metastore.proto
SWAGGER_DIR = swagger
SWAGGER = $(SWAGGER_DIR)/metastore.swagger.json
GOMETASTORE = gometastore
GOPROTO = $(GOMETASTORE)/protobuf

INCLUDES = -I protobuf
INCLUDES += -I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
INCLUDES += -I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway

all: api doc protobuf
	cd $(GOMETASTORE)/hmsv2server && go build
	cd $(GOMETASTORE)/hmsproxy && go build

stats:
	@cloc --no-autogen --git master

api: $(SWAGGER)
	@protoc $(INCLUDES) \
      --swagger_out=logtostderr=true:$(SWAGGER_DIR) \
      $(PROTO)

$(SWAGGER): $(PROTO)

doc: $(PROTO)
	protoc $(INCLUDES) \
        --doc_out=doc --doc_opt=markdown,README.md \
        $(PROTO)
    protoc ${INCLUDES} \
      --doc_out=doc --doc_opt=html,index.html \
      $(PROTO)

protobuf: $(SWAGGER)
	protoc $(INCLUDES) $(PROTO) --go_out=plugins=grpc:$(GOPROTO) && \
    protoc $(INCLUDES) --grpc-gateway_out=logtostderr=true:$(GOPROTO) ${PROTO}

install:
	go get github.com/akolb1/hmsv2api/gometastore/...