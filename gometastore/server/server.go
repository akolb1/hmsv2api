package main

import (
    "flag"
    "net"
    "fmt"
    "log"
    "google.golang.org/grpc"
    pb "github.com/akolb1/hmsv2api/gometastore/protobuf")

var (
    port       = flag.Int("port", 10000, "The server port")
)

type metastoreServer struct {}

func newServer() *metastoreServer {
    return &metastoreServer{}
}

func main() {
    flag.Parse()
    lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    var opts []grpc.ServerOption
    grpcServer := grpc.NewServer(opts...)
    pb.RegisterMetastoreServer(grpcServer, newServer())
    grpcServer.Serve(lis)
}
