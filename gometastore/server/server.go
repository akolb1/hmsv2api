package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"math/rand"
	"time"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/oklog/ulid"
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
	log.Println("CreateDabatase:", req)
	if req.Database == nil || req.Database.Id == nil {
		return nil, fmt.Errorf("missing Database info")
	}
	namespace := req.Database.Id.Namespace
	if namespace == "" {
		return nil, fmt.Errorf("missing namespace")
	}
	dbName := req.Database.Id.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	database := req.Database
	// Create unique ID if it isn's specified
	if database.Id.Id == "" {
		database.Id.Id = getULID()
	}

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
		err = bucket.Put([]byte(dbName), data)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("failed to create database:", err)
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
	log.Println("GetDatabase:", req)
	if req.Id == nil  {
		return nil, fmt.Errorf("missing identity info")
	}
	namespace := req.Id.Namespace
	if namespace == "" {
		return nil, fmt.Errorf("missing namespace")
	}
	dbName := req.Id.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	var database pb.Database
	bucketName := []byte(namespace)

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return fmt.Errorf("bucket %s doesn't exist", bucketName)
		}
		data := bucket.Get([]byte(dbName))
		if data == nil {
			return fmt.Errorf("database %s doesn't exist", dbName)
		}
		err := proto.Unmarshal(data, &database)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("failed to get database:", err)
		return &pb.GetDatabaseResponse{
			Status: &pb.RequestStatus{Status: pb.RequestStatus_ERROR, Error: err.Error()},
		}, nil
	}

	return &pb.GetDatabaseResponse{
		Status:   &pb.RequestStatus{Status: pb.RequestStatus_OK},
		Database: &database,
	}, nil
}

func (s *metastoreServer) ListDatabases(req *pb.ListDatabasesRequest,
	stream pb.Metastore_ListDatabasesServer) error {
	log.Println("ListDatabases", req)
	namespace := req.Namespace
	if namespace == "" {
		return fmt.Errorf("empty namespace")
	}
	bucketName := []byte(namespace)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %s doesn't exist", bucketName)
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			database := new(pb.Database)
			err := proto.Unmarshal(v, database)
			if err != nil {
				continue
			}
			log.Println("send", database)
			if err = stream.Send(database); err != nil {
				log.Println("err sending ", err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println("failed to list databasew:", err)
		return err
	}

	return nil
}

func (s *metastoreServer) DropDatabase(c context.Context,
	req *pb.DropDatabaseRequest) (*pb.RequestStatus, error) {
	log.Println("DropDatabase:", req)
	namespace := req.Id.Namespace
	if namespace == "" {
		return nil, fmt.Errorf("missing empty namespace")
	}
	dbName := req.Id.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	bucketName := []byte(namespace)

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %s doesn't exist", bucketName)
		}
		if data := b.Get([]byte(dbName)); data == nil {
			return fmt.Errorf("database %s doesn't exist", dbName)
		}
		err := b.Delete([]byte(dbName))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("failed to delete database:", err)
		return nil, err
	}

	return &pb.RequestStatus{Status: pb.RequestStatus_OK}, nil
}

// getULID returns a unique ID.
func getULID() string {
	t := time.Unix(1000000, 0)
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
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
