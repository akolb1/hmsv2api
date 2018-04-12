#!/usr/bin/env bash

myhome=$(git rev-parse --show-toplevel)
TOP=${myhome}

GOPATH=~/go
INCLUDES="-I protobuf"
INCLUDES+=" -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis"
INCLUDES+=" -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway"


# Generate Go stubs
cd ${TOP} && protoc \
    ${INCLUDES} \
    protobuf/metastore.proto --go_out=plugins=grpc:gometastore/protobuf

# Generate reverse proxy
cd ${TOP} && protoc \
      ${INCLUDES} \
      --grpc-gateway_out=logtostderr=true:gometastore/protobuf \
      protobuf/metastore.proto

# Generate Swagger
cd ${TOP} && protoc \
    ${INCLUDES} \
    --swagger_out=logtostderr=true:swagger \
    protobuf/metastore.proto

# Generate Python stubs
cd ${TOP} &&
  python -m grpc_tools.protoc \
  ${INCLUDES} \
  --python_out=python/protobuf \
  --grpc_python_out=python/protobuf \
  protobuf/metastore.proto

# Generate Docs
cd ${TOP} && protoc \
  ${INCLUDES} \
    --doc_out=doc --doc_opt=markdown,README.md \
    protobuf/metastore.proto

cd ${TOP} && protoc \
  ${INCLUDES} \
    --doc_out=doc --doc_opt=html,index.html \
    protobuf/metastore.proto