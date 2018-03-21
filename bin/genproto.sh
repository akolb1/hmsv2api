#!/usr/bin/env bash

TOP=$(dirname ${BASH_SOURCE[0]})/..

cd ${TOP} &&
    protoc protobuf/metastore.proto --go_out=plugins=grpc:gometastore


cd ${TOP} &&
  python -m grpc_tools.protoc \
  -I protobuf/ \
  --python_out=python/protobuf \
  --grpc_python_out=python/protobuf \
  protobuf/metastore.proto