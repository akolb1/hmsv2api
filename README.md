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
  - [Markdown](doc/README.md)
  - [HTML](doc/index.html)
  - [Definition](protobuf/metastore.proto)
  - [Swagger/JSON](swagger/metastore.swagger.json)
  - [Swagger/YAML](swagger/swagger.yaml)

## Installation

    go get go get -v github.com/akolb1/hmsv2api/gometastore/...
    
## Running server and proxy

    $ ./gometastore/hmsv2server/hmsv2server -h
    Usage of ./gometastore/hmsv2server/hmsv2server:
      -dbname string
            db name (default "hms2.db")
      -port int
            The server port (default 10010)
            
    $ ./gometastore/hmsproxy/hmsproxy -h
    Usage of ./gometastore/hmsproxy/hmsproxy:
      -hms string
            HMS endpoint (default "localhost:10010")
      -proxy string
            Proxy endpoint (default "localhost:8080")


        
## Updating protobuf definition

The definition is in [protobuf/metastore.proto](protobuf/metastore.proto).
After any changes, please run

    bin/genproto.sh
    
to regenerate all auto-generated files. 