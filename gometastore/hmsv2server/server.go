// Server implementation
package main

import (
	"context"
	"fmt"
	"log"

	"math/rand"
	"time"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/oklog/ulid"
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
	catalog := req.Database.Id.Catalog
	if catalog == "" {
		return nil, fmt.Errorf("missing catalog")
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
		bucket, err := tx.CreateBucketIfNotExists([]byte(catalog))
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
	if req.Id == nil {
		return nil, fmt.Errorf("missing identity info")
	}
	catalog := req.Id.Catalog
	if catalog == "" {
		return nil, fmt.Errorf("missing catalog")
	}
	dbName := req.Id.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	var database pb.Database
	bucketName := []byte(catalog)

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
	catalog := req.Catalog
	if catalog == "" {
		return fmt.Errorf("empty catalog")
	}
	bucketName := []byte(catalog)

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
			log.Println("send", database.Id.Name)
			// Database catalog should match request, no need to send it
			database.Id.Catalog = ""
			// Do not send parameters if not asked to
			if req.ExcludeParams {
				database.Parameters = nil
				database.SystemParameters = nil
			}
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
	catalog := req.Id.Catalog
	if catalog == "" {
		return nil, fmt.Errorf("missing catalog")
	}
	dbName := req.Id.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	bucketName := []byte(catalog)

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
