package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var (
	port       = flag.Int("port", 10000, "The server port")
	boltDbName = flag.String("dbname", "hms2.db", "db name")
)

type metastoreServer struct {
	db *bolt.DB
}

func newServer(db *bolt.DB) *metastoreServer {
	return &metastoreServer{db: db}
}

func (s *metastoreServer) CreateDabatase(c context.Context,
	req *pb.CreateDatabaseRequest) (*pb.GetDatabaseResponse, error) {
	log.Println(req, "session =", req.Cookie.Cookie)
	namespace := req.Database.Id.Namespace.Name
	if namespace == "" {
		return nil, fmt.Errorf("missing empty namespace")
	}
	dbName := req.Database.Id.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	database := req.Database
	data, err := proto.Marshal(database)
	if err != nil {
		log.Println("failed to deserialize", err)
		return nil, err
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		bucket.Put([]byte(dbName), data)
		return nil
	})

	if err != nil {
		return &pb.GetDatabaseResponse{
			Status: &pb.RequestStatus{Status: pb.RequestStatus_ERROR, Error: err.Error()},
		}, nil
	}

	return &pb.GetDatabaseResponse{
		Status:   &pb.RequestStatus{Status: pb.RequestStatus_OK},
		Database: database,
	}, nil
}

func (s *metastoreServer) GetDatabase(c context.Context,
	req *pb.GetDatabaseRequest) (*pb.GetDatabaseResponse, error) {
	log.Println(req, "session =", req.Cookie.Cookie)
	namespace := req.Id.Namespace.Name
	if namespace == "" {
		return nil, fmt.Errorf("missing empty namespace")
	}
	dbName := req.Id.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	var database pb.Database
	bucketName := []byte(namespace)

	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketName).Get([]byte(dbName))
		err := proto.Unmarshal(data, &database)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &pb.GetDatabaseResponse{
			Status: &pb.RequestStatus{Status: pb.RequestStatus_ERROR, Error: err.Error()},
		}, nil
	}

	return &pb.GetDatabaseResponse{
		Status:   &pb.RequestStatus{Status: pb.RequestStatus_OK},
		Database: &database,
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
