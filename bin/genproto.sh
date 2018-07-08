#!/usr/bin/env bash

# This is obsolete, use makefile instead

myhome=$(git rev-parse --show-toplevel)
TOP=${myhome}
PR=protobuf/metastore.proto

GOPATH=~/go
INCLUDES="-I protobuf"
INCLUDES+=" -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis"
INCLUDES+=" -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway"


# Generate Go stubs
cd ${TOP} && protoc \
    ${INCLUDES} \
    ${PROTO} --go_out=plugins=grpc:gometastore/protobuf

# Generate reverse proxy
cd ${TOP} && protoc \
      ${INCLUDES} \
      --grpc-gateway_out=logtostderr=true:gometastore/protobuf \
      ${PROTO}

# Generate Swagger
cd ${TOP} && protoc \
    ${INCLUDES} \
    --swagger_out=logtostderr=true:swagger \
    ${PROTO}

# Generate Python stubs
cd ${TOP} &&
  python -m grpc_tools.protoc \
  ${INCLUDES} \
  --python_out=python/protobuf \
  --grpc_python_out=python/protobuf \
  ${PROTO}

# Generate Docs
cd ${TOP} && protoc \
  ${INCLUDES} \
    --doc_out=docs --doc_opt=markdown,README_proto.md \
    ${PROTO}

cd ${TOP} && protoc \
  ${INCLUDES} \
    --doc_out=docs --doc_opt=html,proto_index.html \
    ${PROTO}