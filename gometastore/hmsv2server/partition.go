package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/boltdb/bolt"
    "github.com/golang/protobuf/proto"
)

func (s *metastoreServer) AddPartition(c context.Context,
	req *pb.AddPartitionRequest) (*pb.RequestStatus, error) {
	log.Println("AddPartition:", req)
	catalog := req.Catalog
	if catalog == "" {
		return nil, fmt.Errorf("missing catalog")
	}
	if req.DbId == nil {
		return nil, fmt.Errorf("missing Db info")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	if req.TableId == nil {
		return nil, fmt.Errorf("missing table info")
	}
	tableName := req.TableId.Name
	if tableName == "" {
		return nil, fmt.Errorf("missing table name")
	}
	partition := req.Partition
	if partition == nil {
		return nil, fmt.Errorf("missing partition data")
	}

	// Construct partition name from values
	values := strings.Join(partition.GetValues(), "/")
	if values == "" {
		return nil, fmt.Errorf("missing partition values")
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		byNameBucket := dbBucket.Bucket([]byte(bynameHdr))
		if byNameBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYNAME info", catalog, dbName)
		}
        tblIdBytes := byNameBucket.Get([]byte(tableName))
        if tblIdBytes == nil {
            return fmt.Errorf("table %s:%s.%s does not exist", catalog, dbName, tableName)
        }
        tablesBucket := dbBucket.Bucket([]byte(tblsHdr))
        if tablesBucket == nil {
            return fmt.Errorf("corrupt catalog %s/%s: no TBLS info", catalog, dbName)
        }
        // Do we have this partition?
        if p := tablesBucket.Get([]byte(values)); p != nil {
            return fmt.Errorf("partition %s already exists", values)
        }
        data, err := proto.Marshal(partition)
        if err != nil {
            return err
        }
        err = tablesBucket.Put([]byte(values), data)
        if err != nil {
            return err
        }

		return nil
	})

	if err != nil {
		log.Println("failed to create table:", err)
		return &pb.RequestStatus{Status: pb.RequestStatus_STATUS_ERROR, Error: err.Error()}, nil
	}

	return &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK}, nil
}
