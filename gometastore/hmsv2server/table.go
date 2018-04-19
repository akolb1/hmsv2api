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
	"fmt"
	"log"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/boltdb/bolt"

	"context"

	"github.com/golang/protobuf/proto"
)

func (s *metastoreServer) CreateTable(c context.Context,
	req *pb.CreateTableRequest) (*pb.GetTableResponse, error) {
	log.Println("CreateTable:", req)
	if req.Table == nil || req.Table.Id == nil {
		return nil, fmt.Errorf("missing Table info")
	}
	if req.DbId == nil {
		return nil, fmt.Errorf("missing Db info")
	}
	catalog := req.Catalog
	if catalog == "" {
		return nil, fmt.Errorf("missing catalog")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	table := req.Table
	if table == nil {
		return nil, fmt.Errorf("missing table data")
	}
	if table.Id == nil {
		return nil, fmt.Errorf("missing table ID")
	}
	tableName := table.Id.Name
	if tableName == "" {
		return nil, fmt.Errorf("missing table name")
	}
	table.Id.Id = getULID()
	id := table.Id.Id
	log.Println("Generated id", id)
    // TODO: Remove compat mode for location
    if table.Location == "" && table.Sd != nil {
        table.Location = table.Sd.Location
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
		byIdBucket := dbBucket.Bucket([]byte(byIDHdr))
		if byIdBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYID info", catalog, dbName)
		}
		tblIdBytes := byNameBucket.Get([]byte(tableName))
		if tblIdBytes != nil {
			return fmt.Errorf("table %s:%s.%s already exists", catalog, dbName, tableName)
		}
		err = byNameBucket.Put([]byte(tableName), []byte(id))
		if err != nil {
			return err
		}
		// Assign unique per-catalog ID
		table.SeqId, _ = dbBucket.NextSequence()

		data, err := proto.Marshal(table)
		if err != nil {
			return err
		}
		err = byIdBucket.Put([]byte(id), data)
		if err != nil {
			return err
		}

		log.Println(id, "->", table)

		return nil
	})

	if err != nil {
		log.Println("failed to create table:", err)
		return &pb.GetTableResponse{
			Status: &pb.RequestStatus{Status: pb.RequestStatus_STATUS_ERROR, Error: err.Error()},
		}, nil
	}

	return &pb.GetTableResponse{
		Status: &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK},
		Table:  table,
	}, nil
}

func (s *metastoreServer) GetTable(c context.Context,
	req *pb.GetTableRequest) (*pb.GetTableResponse, error) {
	log.Println("GetTable:", req)

	if req.Id == nil {
		return nil, fmt.Errorf("missing identity info")
	}
	if req.DbId == nil {
		return nil, fmt.Errorf("missing DB info")
	}
	tableName := req.Id.Name
	if tableName == "" {
		return nil, fmt.Errorf("missing table name")
	}
	catalog := req.Catalog
	if catalog == "" {
		return nil, fmt.Errorf("missing catalog")
	}
	if req.DbId.Name == "" && req.DbId.Id == "" {
		return nil, fmt.Errorf("empty DB info")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing db name")
	}

	var table pb.Table

	err := s.db.View(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		byNameBucket := dbBucket.Bucket([]byte(bynameHdr))
		if byNameBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYNAME info", catalog, dbName)
		}
		byIdBucket := dbBucket.Bucket([]byte(byIDHdr))
		if byIdBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYID info", catalog, dbName)
		}
		tblIdBytes := byNameBucket.Get([]byte(tableName))
		if tblIdBytes == nil {
			return fmt.Errorf("table %s:%s.%s does not exist", catalog, dbName, tableName)
		}
		data := byIdBucket.Get(tblIdBytes)
		if data == nil {
			return fmt.Errorf("catalog corrupted: table %s:%s.%s does not exist",
				catalog, dbName, tableName)
		}
		err = proto.Unmarshal(data, &table)
		if err != nil {
            return fmt.Errorf("catalog corruted: can't decode table data for %s.%s: %v",
                dbName, tableName, err)
        }
        // TODO: Remove compat mode for location
        if table.Location == "" && table.Sd != nil {
            table.Location = table.Sd.Location
        }

		return nil
	})

	if err != nil {
		log.Println("failed to get table:", err)
		return &pb.GetTableResponse{
			Status: &pb.RequestStatus{Status: pb.RequestStatus_STATUS_ERROR, Error: err.Error()},
		}, nil
	}

	return &pb.GetTableResponse{
		Status: &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK},
		Table:  &table,
	}, nil
}

func (s *metastoreServer) ListTables(req *pb.ListTablesRequest,
	stream pb.Metastore_ListTablesServer) error {
	log.Println("ListTables", req)
	if req.DbId == nil {
		return fmt.Errorf("Missing db ID")
	}
	catalog := req.Catalog
	if catalog == "" {
		return fmt.Errorf("missing catalog")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return fmt.Errorf("missing db name")
	}

	err := s.db.View(func(tx *bolt.Tx) error {
		dbBucket, err := getDatabaseBucket(tx, catalog, req.DbId)
		if err != nil {
			return err
		}
		byIdBucket := dbBucket.Bucket([]byte(byIDHdr))
		if byIdBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYNAME info", catalog, dbName)
		}
		byIdBucket.ForEach(func(k, v []byte) error {
			table := new(pb.Table)
			if err := proto.Unmarshal(v, table); err != nil {
				return err
			}

			if len(req.GetFields()) != 0 {
				// Only include specified fields
				tbl := &pb.Table{}
				for _, name := range req.GetFields() {
					switch name {
					case "id.name":
						if tbl.Id == nil {
							tbl.Id = &pb.Id{Name: table.Id.Name}
						} else {
							tbl.Id.Name = table.Id.Name
						}
					case "id":
						tbl.Id = table.Id
					case "location":
					    tbl.Location = table.Location
					    // TODO: Remove compat handling of locaiton
					    if tbl.Location == "" && table.Sd != nil {
					        tbl.Location = table.Sd.Location
                        }
					case "parameters":
						tbl.Parameters = table.Parameters
					case "partkeys":
						tbl.PartitionKeys = table.PartitionKeys
					}
				}
				log.Println("send", tbl)
				if err := stream.Send(tbl); err != nil {
					log.Println("err sending ", err)
					return err
				}
			} else {
                // TODO: Remove compat mode for location
                if table.Location == "" && table.Sd != nil {
                    table.Location = table.Sd.Location
                }
                if err := stream.Send(table); err != nil {
					log.Println("err sending ", err)
					return err
				}
			}
			return nil
		})

		return nil
	})

	if err != nil {
		log.Println("failed to list tables:", err)
		return err
	}

	return nil
}

func (s *metastoreServer) DropTable(c context.Context,
	req *pb.DropTableRequest) (*pb.RequestStatus, error) {
	log.Println("DropTable:", req)
	if req.Id == nil {
		return nil, fmt.Errorf("missing identity info")
	}
	if req.DbId == nil {
		return nil, fmt.Errorf("missing DB info")
	}
	catalog := req.Catalog
	if catalog == "" {
		return nil, fmt.Errorf("missing catalog")
	}
	dbName := req.DbId.Name
	if dbName == "" {
		return nil, fmt.Errorf("missing database name")
	}
	tableName := req.Id.Name
	if tableName == "" {
		return nil, fmt.Errorf("missing table name")
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
		byIdBucket := dbBucket.Bucket([]byte(byIDHdr))
		if byIdBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYID info", catalog, dbName)
		}
		tblIdBytes := byNameBucket.Get([]byte(tableName))
		if tblIdBytes == nil {
			return fmt.Errorf("table %s:%s.%s does not exist", catalog, dbName, tableName)
		}
		tablesBucket := dbBucket.Bucket([]byte(tblsHdr))
		if tablesBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no table info", catalog, dbName)
		}
		tablesBucket.DeleteBucket(tblIdBytes)
		byIdBucket.Delete(tblIdBytes)
		byNameBucket.Delete([]byte(tableName))
		return nil
	})

	if err != nil {
		log.Println("failed to delete table:", err)
		return nil, err
	}

	return &pb.RequestStatus{Status: pb.RequestStatus_STATUS_OK}, nil
}
