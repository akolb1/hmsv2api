package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/boltdb/bolt"
	"google.golang.org/grpc"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
)

var (
	port       = flag.Int("port", 10000, "The server port")
	boltDbName = flag.String("dbname", "hms2.db", "db name")
)

func main() {
	flag.Parse()
	db, err := bolt.Open(*boltDbName, 0644, nil)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	defer db.Close()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMetastoreServer(grpcServer, newServer(db))
	grpcServer.Serve(lis)
}
