#

PROTOC = protoc
PROTO = protobuf/metastore.proto
SWAGGER_DIR = swagger
SWAGGER = $(SWAGGER_DIR)/metastore.swagger.json
GOMETASTORE = gometastore
GOPROTO = $(GOMETASTORE)/protobuf
GITHUB = github.com
THIS = $(GITHUB)/akolb1/hmsv2api
GO_ALL = $(THIS)/$(GOMETASTORE)/...


INCLUDES = -I protobuf
INCLUDES += -I $(GOPATH)/src/$(GITHUB)/grpc-ecosystem/grpc-gateway/third_party/googleapis
INCLUDES += -I $(GOPATH)/src/$(GITHUB)/grpc-ecosystem/grpc-gateway

all: build

build:
	cd $(GOMETASTORE)/hmsv2server && go build
	cd $(GOMETASTORE)/hmsproxy && go build

stats:
	@cloc --no-autogen --git master

api: $(SWAGGER)
	@$(PROTOC) $(INCLUDES) \
      --swagger_out=logtostderr=true:$(SWAGGER_DIR) \
      $(PROTO)

$(SWAGGER): $(PROTO)

docs: $(PROTO)
	$(PROTOC) $(INCLUDES) \
        --doc_out=docs --doc_opt=markdown,README.md \
        $(PROTO)
	$(PROTOC) ${INCLUDES} \
      --doc_out=docs --doc_opt=html,index.html \
      $(PROTO)

deps:
	go get $(GITHUB)/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
	go get $(GITHUB)/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get $(GITHUB)/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get $(GITHUB)/golang/protobuf/protoc-gen-go

proto:
	@ if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	go generate $(GO_ALL)

install:
	go get $(GOALL)
