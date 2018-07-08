# gRPC definition for Metadata API

The file [metastore.proto](metastore.proto) contains definition of the metadata API.

It uses extensions that are not commonly installed:

- [google/api/annotations.proto](https://github.com/googleapis/googleapis/blob/master/google/api/annotations.proto)
- [protoc-gen-swagger/options/annotations.proto](https://github.com/grpc-ecosystem/grpc-gateway/blob/master/protoc-gen-swagger/options/annotations.proto)

So this file may need to be edited if it is used to generate
Java code. 

Whenever any changes are made in the proto file, you should run

    make proto
     
in the top level directory.