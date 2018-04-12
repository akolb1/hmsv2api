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
