#!/usr/bin/env bash

cd $( dirname "${BASH_SOURCE[0]}" )/.. &&
    protoc protobuf/metastore.proto --go_out=plugins=grpc:gometastore