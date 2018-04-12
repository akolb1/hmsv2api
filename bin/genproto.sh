#!/usr/bin/env bash

myhome=$(git rev-parse --show-toplevel)
TOP=${myhome}

GOPATH=~/go

# Generate Go stubs
cd ${TOP} && protoc \
      -I protobuf \
      -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    protobuf/metastore.proto --go_out=plugins=grpc:gometastore/protobuf

# Generate reverse proxy
cd ${TOP} && protoc \
      -I protobuf \
      -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
      --grpc-gateway_out=logtostderr=true:gometastore/protobuf \
      protobuf/metastore.proto

# Generate Swagger
cd ${TOP} && protoc \
    -I protobuf \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --swagger_out=logtostderr=true:swagger \
    protobuf/metastore.proto

# Generate Python stubs
cd ${TOP} &&
  python -m grpc_tools.protoc \
  -I protobuf \
  -I ${GOPATH}/src \
  -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --python_out=python/protobuf \
  --grpc_python_out=python/protobuf \
  protobuf/metastore.proto

# Generate Docs
cd ${TOP} && protoc \
    -I protobuf \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --doc_out=doc --doc_opt=markdown,README.md \
    protobuf/metastore.proto

cd ${TOP} && protoc \
    -I protobuf \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --doc_out=doc --doc_opt=html,index.html \
    protobuf/metastore.proto