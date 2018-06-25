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

	partition.Id.Id = getULID()

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
		log.Println("added partition", partition)

		return nil
	})

	if err != nil {
		log.Println("failed to create partition:", err)
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

		var table pb.Table

		// Extract table
		byNameBucket := dbBucket.Bucket([]byte(bynameHdr))
		if byNameBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYNAME info", catalog, dbName)
		}
		byIDBucket := dbBucket.Bucket([]byte(byIDHdr))
		if byIDBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYID info", catalog, dbName)
		}
		tblIDBytes := byNameBucket.Get([]byte(tableName))
		if tblIDBytes == nil {
			return fmt.Errorf("table %s:%s.%s does not exist", catalog, dbName, tableName)
		}
		tableData := byIDBucket.Get(tblIDBytes)
		if tableData == nil {
			return fmt.Errorf("catalog corrupted: table %s:%s.%s does not exist",
				catalog, dbName, tableName)
		}
		if err = proto.Unmarshal(tableData, &table); err != nil {
			return fmt.Errorf("catalog corruted: can't decode table data for %s.%s: %v",
				dbName, tableName, err)
		}

		// Do we have this partition?
		data := tablesBucket.Get([]byte(values))
		if data == nil {
			return fmt.Errorf("no partition %s.%s/%s", dbName, tableName, values)
		}
		if err := proto.Unmarshal(data, &partition); err != nil {
			return err
		}
		partition.Table = &table
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

	// Values that we are interested in
	valuesMap := make(map[string]bool)

	if req.Values != nil {
		for _, values := range req.Values {
			vString := strings.Join(values.GetValue(), "/")
			valuesMap[vString] = true
		}
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

		// Extract table
		byNameBucket := dbBucket.Bucket([]byte(bynameHdr))
		if byNameBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYNAME info", catalog, dbName)
		}
		byIDBucket := dbBucket.Bucket([]byte(byIDHdr))
		if byIDBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYID info", catalog, dbName)
		}
		tblIDBytes := byNameBucket.Get([]byte(tableName))
		if tblIDBytes == nil {
			return fmt.Errorf("table %s:%s.%s does not exist", catalog, dbName, tableName)
		}
		data := byIDBucket.Get(tblIDBytes)
		if data == nil {
			return fmt.Errorf("catalog corrupted: table %s:%s.%s does not exist",
				catalog, dbName, tableName)
		}

		var table pb.Table
		err = proto.Unmarshal(data, &table)
		if err != nil {
			return fmt.Errorf("catalog corruted: can't decode table data for %s.%s: %v",
				dbName, tableName, err)
		}

		first := true

		walker := func(k, v []byte) error {
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
					case "parameters":
						part.Parameters = partition.Parameters
					case "values":
						part.Values = partition.Values
					case "sd":
						part.Sd = partition.Sd
					case "sd.parameters":
						if part.Sd == nil {
							part.Sd = &pb.StorageDescriptor{Parameters: partition.Sd.Parameters}
							log.Println("sd = ", part.Sd, partition.Sd.Parameters)
						} else {
							part.Sd.Parameters = partition.Sd.Parameters
						}
					case "sd.serdeinfo":
						if part.Sd == nil {
							part.Sd = &pb.StorageDescriptor{SerdeInfo: partition.Sd.SerdeInfo}
						} else {
							part.Sd.SerdeInfo = partition.Sd.SerdeInfo
						}
					case "sd.serdeinfo.parameters":
						if part.Sd == nil {
							part.Sd = &pb.StorageDescriptor{SerdeInfo: &pb.SerDeInfo{Parameters: partition.Sd.SerdeInfo.Parameters}}
						} else {
							part.Sd.SerdeInfo.Parameters = partition.Sd.SerdeInfo.Parameters
						}
					case "table":
						if first {
							// Include table in first partition only
							first = false
							part.Table = &table
						}
					}
				}

				if req.GetExclude() != nil {
				    excludeParts(part, req.GetExclude())
                }
				log.Println("send", part)
				if err := stream.Send(part); err != nil {
					log.Println("err sending:", err)
					return err
				}
			} else {
				if first {
					// Include table in first partition only
					first = false
					partition.Table = &table
				}
				log.Println("Send partition", partition)
                if req.GetExclude() != nil {
                    excludeParts(partition, req.GetExclude())
                }
				if err := stream.Send(partition); err != nil {
					log.Println("err sending:", err)
					return err
				}
			}
			return nil
		}

		if len(valuesMap) == 0 {
			tablesBucket.ForEach(walker)
		} else {
			// Walk over values only
			for value := range valuesMap {
				v := tablesBucket.Get([]byte(value))
				if v == nil {
					continue
				}
				walker([]byte(value), v)
			}
		}

		return nil
	})

	if err != nil {
		log.Println("failed to list partitions:", err)
		return err
	}

	return nil
}

func (s *metastoreServer) DropPartitions(c context.Context,
	req *pb.DropPartitionsRequest) (*pb.RequestStatus, error) {
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
	partitionValues := req.GetValues()

	err := s.db.Update(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		tablesBucket, err := getTableBucket(dbBucket, catalog, dbName, tableName, false)
		if err != nil {
			return err
		}
		for _, values := range partitionValues {
			vString := strings.Join(values.GetValue(), "/")
			if err = tablesBucket.Delete([]byte(vString)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println("failed to delete table:", err)
		return nil, err
	}

	return &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK}, nil
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

func excludeParts(part *pb.Partition, exclude []string) {
	for _, e := range exclude {
		switch e {
		case "location":
			part.Location = ""
		case "parameters":
			part.Parameters = nil
		case "values":
			part.Values = nil
		case "sd":
			part.Sd = nil
		case "sd.parameters":
			if part.Sd != nil {
				part.Sd.Parameters = nil
			}
		case "sd.serdeinfo":
			if part.Sd != nil {
				part.Sd.SerdeInfo = nil
			}
		case "sd.serdeinfo.parameters":
			if part.Sd != nil {
				part.Sd.Parameters = nil
			}
		case "table":
			part.Table = nil
		}
	}
}
