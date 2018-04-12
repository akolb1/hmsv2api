package main

import (
    "log"
    "fmt"
    "github.com/boltdb/bolt"
    "github.com/golang/protobuf/proto"

    pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
    "context"
)

func (s *metastoreServer) CreateDabatase(c context.Context,
    req *pb.CreateDatabaseRequest) (*pb.GetDatabaseResponse, error) {
    log.Println("CreateDabatase:", req)
    if req.Database == nil || req.Database.Id == nil {
        return nil, fmt.Errorf("missing Database info")
    }
    catalog := req.Catalog
    if catalog == "" {
        return nil, fmt.Errorf("missing catalog")
    }
    dbName := req.Database.Id.Name
    if dbName == "" {
        return nil, fmt.Errorf("missing database name")
    }
    database := req.Database
    // Create unique ID if it isn's specified
    database.Id.Id = getULID()
    id := database.Id.Id

    err := s.db.Update(func(tx *bolt.Tx) error {
        catBucket, err := tx.CreateBucketIfNotExists([]byte(catalog))
        if err != nil {
            return err
        }
        nameMap, err := catBucket.CreateBucketIfNotExists([]byte(bynameHdr))
        if err != nil {
            return err
        }
        // Do we have DB with this name?
        if r := nameMap.Get([]byte(dbName)); r != nil {
            // Table alreday exists
            return fmt.Errorf("database %s already exists", dbName)
        }

        idMap, err := catBucket.CreateBucketIfNotExists([]byte(byIDHdr))
        if err != nil {
            return err
        }
        dbBucket, err := catBucket.CreateBucketIfNotExists([]byte(dbHdr))
        if err != nil {
            return err
        }
        // Create structure for the DB needed for tables
        dbDataBucket, err := dbBucket.CreateBucketIfNotExists([]byte(id))
        if err != nil {
            return err
        }
        _, err = dbDataBucket.CreateBucketIfNotExists([]byte(bynameHdr))
        if err != nil {
            return err
        }
        _, err = dbDataBucket.CreateBucketIfNotExists([]byte(byIDHdr))
        if err != nil {
            return err
        }
        _, err = dbDataBucket.CreateBucketIfNotExists([]byte(tblsHdr))
        if err != nil {
            return err
        }

        // Put mapping of name to ID
        err = nameMap.Put([]byte(dbName), []byte(id))
        if err != nil {
            return err
        }

        // Assign unique per-catalog ID
        database.SeqId, _ = catBucket.NextSequence()

        // Store database info in idMap
        data, err := proto.Marshal(database)
        if err != nil {
            log.Println("failed to deserialize", err)
            return err
        }

        err = idMap.Put([]byte(id), data)
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
    catalog := req.Catalog
    if catalog == "" {
        return nil, fmt.Errorf("missing catalog")
    }
    dbName := req.Id.Name
    id := req.Id.Id
    if dbName == "" && id == "" {
        return nil, fmt.Errorf("missing database name or id")
    }

    var database pb.Database
    bucketName := []byte(catalog)
    if err := s.db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte(catalog))
        if err != nil {
            return err
        }
        return nil
    }); err != nil {
        return nil, err
    }

    err := s.db.View(func(tx *bolt.Tx) error {
        catalogBucket := tx.Bucket(bucketName)
        if catalogBucket == nil {
            return fmt.Errorf("bucket %s doesn't exist", catalog)
        }
        idBucket := catalogBucket.Bucket([]byte(byIDHdr))
        if idBucket == nil {
            return fmt.Errorf("corrupt catalog - missing ID map")
        }
        idBytes := []byte(id)
        if id == "" {
            // Locate ID by name
            nameIdBucket := catalogBucket.Bucket([]byte(bynameHdr))
            if nameIdBucket == nil {
                return fmt.Errorf("corrupt catalog - missing NAME map")
            }
            idBytes = nameIdBucket.Get([]byte(dbName))
            if idBytes == nil {
                return fmt.Errorf("database %s doesn't exist", dbName)
            }
        }

        data := idBucket.Get(idBytes)
        if data == nil {
            return fmt.Errorf("corrupt catalog - missing db %s", string(idBytes))
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
    if err := s.db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists(bucketName)
        if err != nil {
            return err
        }
        return nil
    }); err != nil {
        return err
    }

    err := s.db.View(func(tx *bolt.Tx) error {
        catalogBucket := tx.Bucket(bucketName)
        if catalogBucket == nil {
            return fmt.Errorf("bucket %s doesn't exist", bucketName)
        }
        idMap := catalogBucket.Bucket([]byte(byIDHdr))
        if idMap == nil {
            return nil
        }

        idMap.ForEach(func(k, v []byte) error {
            database := new(pb.Database)
            err := proto.Unmarshal(v, database)
            if err != nil {
                return nil
            }
            log.Println("send", database.Id.Name)
            // Do not send parameters if not asked to
            if req.ExcludeParams {
                database.Parameters = nil
                database.SystemParameters = nil
            }
            if err = stream.Send(database); err != nil {
                log.Println("err sending ", err)
                return err
            }
            return nil
        })
        return nil
    })

    if err != nil {
        log.Println("failed to list databases:", err)
        return err
    }

    return nil
}

func (s *metastoreServer) DropDatabase(c context.Context,
    req *pb.DropDatabaseRequest) (*pb.RequestStatus, error) {
    log.Println("DropDatabase:", req)
    if req.Id == nil {
        return nil, fmt.Errorf("missing identity info")
    }
    catalog := req.Catalog
    if catalog == "" {
        return nil, fmt.Errorf("missing catalog")
    }
    dbName := req.Id.Name
    if dbName == "" {
        return nil, fmt.Errorf("missing database name")
    }

    err := s.db.Update(func(tx *bolt.Tx) error {
        catalogBucket := tx.Bucket([]byte(catalog))
        if catalogBucket == nil {
            return fmt.Errorf("bucket %s doesn't exist", catalog)
        }
        idMap := catalogBucket.Bucket([]byte(byIDHdr))
        if idMap == nil {
            return fmt.Errorf("corrupd database: missing ID map")
        }
        nameMap := catalogBucket.Bucket([]byte(bynameHdr))
        if nameMap == nil {
            return fmt.Errorf("corrupd database: missing name map")
        }
        idBytes := nameMap.Get([]byte(dbName))
        if idBytes == nil {
            return fmt.Errorf("database %s doesn't exist", dbName)
        }
        // Remove info from this DB
        nameMap.Delete([]byte(dbName))
        idMap.Delete(idBytes)
        if dbInfo := catalogBucket.Bucket([]byte(dbHdr)); dbInfo != nil {
            dbInfo.DeleteBucket(idBytes)
        }

        return nil
    })

    if err != nil {
        log.Println("failed to delete database:", err)
        return nil, err
    }

    return &pb.RequestStatus{Status: pb.RequestStatus_OK}, nil
}

