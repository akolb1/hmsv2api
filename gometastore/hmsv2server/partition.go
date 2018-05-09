// DB Structure:
//
// root+
//   catalog1+
//           |
//           + BYNAME Name -> Id
//           + BYID   Id -> { Database }
//           + DB +
//                |
//                +<id1>
//                    BYNAME Name -> Id
//                    BYID   ID -> { Table }
//                    TBLS
//                       + <id1>
//                            DATA
//                            PARTS
//                       + <id2>
//                            DATA
//                            PARTS
//                |
//                + <id2>
//                    DATA
//                    TBLS
//

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"io"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
)

// TODO: Figure out schema evolution for partitions

func (s *metastoreServer) AddPartition(c context.Context,
	req *pb.AddPartitionRequest) (*pb.AddPartitionResponse, error) {
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
		tablesBucket, err := getTableBucket(dbBucket, catalog, dbName, tableName, true)
		if err != nil {
			return err
		}
		// Do we have this partition?
		if p := tablesBucket.Get([]byte(values)); p != nil {
			return fmt.Errorf("partition %s already exists", values)
		}
		partition.SeqId, _ = tablesBucket.NextSequence()
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
		return &pb.AddPartitionResponse{
			Sequence: req.Sequence,
			Status: &pb.RequestStatus{
				Status: pb.RequestStatus_STATUS_ERROR,
				Error:  err.Error(),
			},
		}, nil
	}

	return &pb.AddPartitionResponse{
		Sequence: req.Sequence,
		Status:   &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK},
	}, nil
}

func (s *metastoreServer) AddManyPartitions(stream pb.Metastore_AddManyPartitionsServer) error {
	// Read first request
	log.Println("AddMnayPartitions")
	req, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	log.Println("AddPartition:", req)
	catalog := req.Catalog
	if catalog == "" {
		return fmt.Errorf("missing catalog")
	}
	if req.DbId == nil {
		return fmt.Errorf("missing Db info")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return fmt.Errorf("missing database name")
	}
	if req.TableId == nil {
		return fmt.Errorf("missing table info")
	}
	tableName := req.TableId.Name
	if tableName == "" {
		return fmt.Errorf("missing table name")
	}
	partition := req.Partition
	if partition == nil {
		return fmt.Errorf("missing partition data")
	}

	// Construct partition name from values
	values := strings.Join(partition.GetValues(), "/")
	if values == "" {
		return fmt.Errorf("missing partition values")
	}
	return nil
}

func (s *metastoreServer) GetPartition(c context.Context,
	req *pb.GetPartitionRequest) (*pb.GetPartitionResponse, error) {
	log.Println("GetPartition:", req)
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

	// Construct partition name from values
	values := strings.Join(req.GetValues(), "/")
	if values == "" {
		return nil, fmt.Errorf("missing partition values")
	}
	var partition pb.Partition

	err := s.db.View(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		tablesBucket, err := getTableBucket(dbBucket, catalog, dbName, tableName, false)
		if err != nil {
			return err
		}
		// Do we have this partition?
		data := tablesBucket.Get([]byte(values))
		if data == nil {
			return fmt.Errorf("no partition %s.%s/%s", dbName, tableName, values)
		}
		if err := proto.Unmarshal(data, &partition); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("failed to get partition:", err)
		return &pb.GetPartitionResponse{
			Status: &pb.RequestStatus{Status: pb.RequestStatus_STATUS_ERROR, Error: err.Error()},
		}, nil
	}

	return &pb.GetPartitionResponse{
		Status:    &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK},
		Partition: &partition,
	}, nil
}

func (s *metastoreServer) ListPartitions(req *pb.ListPartitionsRequest,
	stream pb.Metastore_ListPartitionsServer) error {
	log.Println("ListPartitions:", req)
	catalog := req.Catalog
	if catalog == "" {
		return fmt.Errorf("missing catalog")
	}
	if req.DbId == nil {
		return fmt.Errorf("missing Db info")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return fmt.Errorf("missing database name")
	}
	if req.TableId == nil {
		return fmt.Errorf("missing table info")
	}
	tableName := req.TableId.Name
	if tableName == "" {
		return fmt.Errorf("missing table name")
	}

	err := s.db.View(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		tablesBucket, err := getTableBucket(dbBucket, catalog, dbName, tableName, false)
		if err != nil {
			return err
		}
		tablesBucket.ForEach(func(k, v []byte) error {
			partition := new(pb.Partition)
			if err := proto.Unmarshal(v, partition); err != nil {
				return nil
			}
			if len(req.GetFields()) != 0 {
				// Only include specified fields
				part := &pb.Partition{}
				for _, name := range req.GetFields() {
					switch name {
					case "location":
						part.Location = partition.Location
						if part.Location == "" && partition.Sd != nil {
							part.Location = partition.Sd.Location
						}
					case "parameters":
						part.Parameters = partition.Parameters
					case "values":
						part.Values = partition.Values
					}
				}
				log.Println("send", part)
				if err := stream.Send(part); err != nil {
					log.Println("err sending:", err)
					return err
				}
			} else {
				if err := stream.Send(partition); err != nil {
					log.Println("err sending:", err)
					return err
				}
			}
			return nil
		})
		return nil
	})

	if err != nil {
		log.Println("failed to list partitions:", err)
		return err
	}

	return nil
}

func (s *metastoreServer) DropPartition(c context.Context,
	req *pb.DropPartitionRequest) (*pb.RequestStatus, error) {
	log.Println("DropPartition:", req)
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
	// Construct partition name from values
	values := strings.Join(req.GetValues(), "/")
	if values == "" {
		return nil, fmt.Errorf("missing partition values")
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		tablesBucket, err := getTableBucket(dbBucket, catalog, dbName, tableName, false)
		if err != nil {
			return err
		}
		if err = tablesBucket.Delete([]byte(values)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("failed to delete table:", err)
		return nil, err
	}

	return &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK}, nil
}

func (s *metastoreServer) DropPartitions(stream pb.Metastore_DropPartitionsServer) error {
	// Read first request
	log.Println("DropPartitions")
	req, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			log.Println("no partitions to drop")
			return nil
		}
		return err
	}

	catalog := req.Catalog
	if catalog == "" {
		return fmt.Errorf("missing catalog")
	}
	if req.DbId == nil {
		return fmt.Errorf("missing Db info")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return fmt.Errorf("missing database name")
	}
	if req.TableId == nil {
		return fmt.Errorf("missing table info")
	}
	tableName := req.TableId.Name
	if tableName == "" {
		return fmt.Errorf("missing table name")
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		tablesBucket, err := getTableBucket(dbBucket, catalog, dbName, tableName, false)
		if err != nil {
			return err
		}
		for {
			values := strings.Join(req.GetValues(), "/")
			if values == "" {
				log.Println("missing value from request")
				continue
			}
			log.Println("dropping ", values)
			if err = tablesBucket.Delete([]byte(values)); err != nil {
				return err
			}
			req, err = stream.Recv()
			if err != nil {
				if err == io.EOF {
					log.Println("done dropping partitions")
					return nil
				}
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println("failed to drop partitions:", err)
		return err
	}

	return nil
}

func getTableBucket(dbBucket *bolt.Bucket, catalog string, dbName string, tableName string,
	create bool) (*bolt.Bucket, error) {
	byNameBucket := dbBucket.Bucket([]byte(bynameHdr))
	if byNameBucket == nil {
		return nil, fmt.Errorf("corrupt catalog %s/%s: no BYNAME info", catalog, dbName)
	}
	tblIDBytes := byNameBucket.Get([]byte(tableName))
	if tblIDBytes == nil {
		return nil, fmt.Errorf("table %s:%s.%s does not exist", catalog, dbName, tableName)
	}
	tablesBucket := dbBucket.Bucket([]byte(tblsHdr))
	if tablesBucket == nil {
		return nil, fmt.Errorf("corrupt catalog %s/%s: no TBLS info", catalog, dbName)
	}
	tBucket := tablesBucket.Bucket(tblIDBytes)
	if tBucket == nil {
		if !create {
			return nil, fmt.Errorf("corrupt catalog %s/%s: no TBLS info", catalog, dbName)
		}
		tBucket, err := tablesBucket.CreateBucketIfNotExists(tblIDBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to create table bucket for %s: %v", tableName, err)
		}
		return tBucket, nil
	}
	return tBucket, nil
}
