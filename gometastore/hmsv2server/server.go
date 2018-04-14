// Server implementation
//
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
	"strings"

	"fmt"

	pb "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/boltdb/bolt"
	"github.com/imdario/go-ulid"
)

const (
	bynameHdr = "BYNAME"
	byIDHdr   = "BYID"
	dbHdr     = "DB"
	tblsHdr   = "TBLS"
)

type metastoreServer struct {
	db *bolt.DB
}

func newServer(db *bolt.DB) *metastoreServer {
	return &metastoreServer{db: db}
}

// Table ops

// getULID returns a unique ID.
func getULID() string {
	return strings.TrimRight(ulid.New().String(), "\u0000")
}

// getDatabaseBucket returns Database bucket for the database specified by ID.
//
//   tx - Bolt transaction
//   catalog - catalog name, must be non-empty
//   db - Database ID, must be non-empty and either name or Id should be specified
func getDatabaseBucket(tx *bolt.Tx, catalog string, db *pb.Id) (bucket *bolt.Bucket, err error) {
	catBucket := tx.Bucket([]byte(catalog))
	if catBucket == nil {
		return nil, fmt.Errorf("missing catalog %s", catalog)
	}
	idMap := catBucket.Bucket([]byte(byIDHdr))
	if idMap == nil {
		return nil, fmt.Errorf("corrupted catalog %s: missing ID map", catalog)
	}
	idBytesDb := []byte(db.Id)
	if db.Id == "" {
		// Locate DB ID by name
		nameIdBucket := catBucket.Bucket([]byte(bynameHdr))
		if nameIdBucket == nil {
			return nil, fmt.Errorf("corrupt catalog - missing NAME map")
		}
		idBytesDb = nameIdBucket.Get([]byte(db.Name))
		if idBytesDb == nil {
			return nil, fmt.Errorf("database %s doesn't exist", db.Name)
		}
	}
	dbInfoBucket := catBucket.Bucket([]byte(dbHdr))
	if dbInfoBucket == nil {
		return nil, fmt.Errorf("corrupt catalog %s: no DB info", catalog)
	}
	dbBucket := dbInfoBucket.Bucket(idBytesDb)
	if dbBucket == nil {
		return nil, fmt.Errorf("corrupt catalog %s/%s: no DB info", catalog, db.Name)
	}

	return dbBucket, nil
}
