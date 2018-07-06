# Hive Metastore API v2

## Introduction

This is work in progress for Hive Metastore API version 2.
It has the following components:

- [API definition](protobuf/metastore.proto)
- Tiny server-side reference [implementation](gometastore/hmsv2server)
- HTTP - gRPC [proxy](gometastore/hmsproxy)
- Automatically produced documentation

All Java work required to use the new API is elsewhere.

## Documentation

- API Documentation
  - [Markdown](https://akolb1.github.io/hmsv2api/README_proto.md)
  - [HTML](https://akolb1.github.io/hmsv2api/proto_index.html)
  - [Definition](protobuf/metastore.proto)
  - [Swagger/JSON](swagger/metastore.swagger.json)
  - [Swagger/YAML](swagger/swagger.yaml)

## Installation

    go get github.com/akolb1/hmsv2api/gometastore/...
    
## Prerequisites

This project uses gRPC version 3 and needs Version 3 protoc compiler in your path.
    
## Running server and proxy

    $ hmsv2server -h
    Usage of ./gometastore/hmsv2server/hmsv2server:
      -dbname string
            db name (default "hms2.db")
      -port int
            The server port (default 10010)
            
    $ hmsproxy -h
    Usage of ./gometastore/hmsproxy/hmsproxy:
      -hms string
            HMS endpoint (default "localhost:10010")
      -proxy string
            Proxy endpoint (default "localhost:8080")


        
## Updating protobuf definition

The definition is in [protobuf/metastore.proto](protobuf/metastore.proto).
After any changes, please run

    make deps # Only need to run this once
    make proto
    make
    
to regenerate all auto-generated files. 
To regenerate documentation use

    make doc