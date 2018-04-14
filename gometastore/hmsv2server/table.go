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
	dbId := req.DbId.Id
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

	err := s.db.Update(func(tx *bolt.Tx) error {
		catBucket := tx.Bucket([]byte(catalog))
		if catBucket == nil {
			return fmt.Errorf("missing catalog %s", catalog)
		}
		idMap := catBucket.Bucket([]byte(byIDHdr))
		if idMap == nil {
			return fmt.Errorf("database %s does not exist in %s", dbName, catalog)
		}

		idBytesDb := []byte(dbId)
		if dbId == "" {
			// Locate DB ID by name
			nameIdBucket := catBucket.Bucket([]byte(bynameHdr))
			if nameIdBucket == nil {
				return fmt.Errorf("corrupt catalog - missing NAME map")
			}
			idBytesDb = nameIdBucket.Get([]byte(dbName))
			if idBytesDb == nil {
				return fmt.Errorf("database %s doesn't exist", dbName)
			}
		}
		dbInfoBucket := catBucket.Bucket([]byte(dbHdr))
		if dbInfoBucket == nil {
			return fmt.Errorf("corrupt catalog %s: no DB info", catalog)
		}
		dbBucket := dbInfoBucket.Bucket(idBytesDb)
		if dbBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no DB info", catalog, dbName)
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
		err := byNameBucket.Put([]byte(tableName), []byte(id))
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
		catalogBucket := tx.Bucket([]byte(catalog))
		if catalogBucket == nil {
			return fmt.Errorf("bucket %s doesn't exist", catalog)
		}

		// Locate ID by name
		nameIdBucket := catalogBucket.Bucket([]byte(bynameHdr))
		if nameIdBucket == nil {
			return fmt.Errorf("corrupt catalog - missing NAME map")
		}
		idBytesDb := nameIdBucket.Get([]byte(dbName))
		if idBytesDb == nil {
			return fmt.Errorf("database %s doesn't exist", dbName)
		}
		dbInfoBucket := catalogBucket.Bucket([]byte(dbHdr))
		if dbInfoBucket == nil {
			return fmt.Errorf("corrupt catalog %s: no DB info", catalog)
		}
		dbBucket := dbInfoBucket.Bucket(idBytesDb)
		if dbBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no DB info", catalog, dbName)
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
		if err := proto.Unmarshal(data, &table); err != nil {
			return err
		} else {
			return nil
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
		catalogBucket := tx.Bucket([]byte(catalog))
		if catalogBucket == nil {
			return fmt.Errorf("bucket %s doesn't exist", catalog)
		}

		// Locate ID by name
		nameIdBucket := catalogBucket.Bucket([]byte(bynameHdr))
		if nameIdBucket == nil {
			return fmt.Errorf("corrupt catalog - missing NAME map")
		}
		idBytesDb := nameIdBucket.Get([]byte(dbName))
		if idBytesDb == nil {
			return fmt.Errorf("database %s doesn't exist", dbName)
		}
		dbInfoBucket := catalogBucket.Bucket([]byte(dbHdr))
		if dbInfoBucket == nil {
			return fmt.Errorf("corrupt catalog %s: no DB info", catalog)
		}
		dbBucket := dbInfoBucket.Bucket(idBytesDb)
		if dbBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no DB info", catalog, dbName)
		}
		byIdBucket := dbBucket.Bucket([]byte(byIDHdr))
		if byIdBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no BYNAME info", catalog, dbName)
		}
		byIdBucket.ForEach(func(k, v []byte) error {
			table := new(pb.Table)
			if err := proto.Unmarshal(v, table); err != nil {
				return nil
			}
			log.Println("send", table.Id.Name)
			// table.Id.Catalog = ""
			if err := stream.Send(table); err != nil {
				log.Println("err sending ", err)
				return err
			}
			return nil
		})

		return nil
	})

	if err != nil {
		log.Println("failed to list tables:", err)
		return err
	}

	return err
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
		catalogBucket := tx.Bucket([]byte(catalog))
		if catalogBucket == nil {
			return fmt.Errorf("bucket %s doesn't exist", catalog)
		}

		// Locate ID by name
		nameIdBucket := catalogBucket.Bucket([]byte(bynameHdr))
		if nameIdBucket == nil {
			return fmt.Errorf("corrupt catalog - missing NAME map")
		}
		idBytesDb := nameIdBucket.Get([]byte(dbName))
		if idBytesDb == nil {
			return fmt.Errorf("database %s doesn't exist", dbName)
		}
		dbInfoBucket := catalogBucket.Bucket([]byte(dbHdr))
		if dbInfoBucket == nil {
			return fmt.Errorf("corrupt catalog %s: no DB info", catalog)
		}
		dbBucket := dbInfoBucket.Bucket(idBytesDb)
		if dbBucket == nil {
			return fmt.Errorf("corrupt catalog %s/%s: no DB info", catalog, dbName)
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
