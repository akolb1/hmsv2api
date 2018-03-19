package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type metastoreServer struct{}

func newServer() *metastoreServer {
	return &metastoreServer{}
}

func (*metastoreServer) CreateDabatase(c context.Context,
	req *pb.CreateDatabaseRequest) (*pb.GetDatabaseResponse, error) {
	log.Println(req)
	id := &pb.Id{Name: "someName", Namespace: &pb.Namespace{"namespace"}}
	return &pb.GetDatabaseResponse{
		Status:   &pb.RequestStatus{Status: pb.RequestStatus_OK},
		Database: &pb.Database{Id: id},
	}, nil
}

func (*metastoreServer) GetDatabase(c context.Context,
	req *pb.GetDatabaseRequest) (*pb.GetDatabaseResponse, error) {
	id := &pb.Id{Name: "someNameOther", Namespace: &pb.Namespace{"namespace"}}
	log.Println(req)
	return &pb.GetDatabaseResponse{
		Status:   &pb.RequestStatus{Status: pb.RequestStatus_OK},
		Database: &pb.Database{Id: id},
	}, nil
}

func (*metastoreServer) ListDatabases(*pb.ListDatabasesRequest,
	pb.Metastore_ListDatabasesServer) error {
	return nil
}

func (*metastoreServer) DropDatabase(context.Context,
	*pb.DropDatabaseRequest) (*pb.RequestStatus, error) {
	return &pb.RequestStatus{Status: pb.RequestStatus_OK}, nil
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
