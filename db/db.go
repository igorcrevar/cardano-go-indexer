package db

import (
	"igorcrevar/cardano-go-syncer/core"
	"igorcrevar/cardano-go-syncer/db/boltdb"
	"igorcrevar/cardano-go-syncer/db/leveldb"
	"strings"
)

func NewDatabase(name string) core.Database {
	switch strings.ToLower(name) {
	case "leveldb":
		return &leveldb.LevelDbDatabase{}
	default:
		return &boltdb.BoltDatabase{}
	}
}

func NewDatabaseInit(name string, filePath string) (core.Database, error) {
	db := NewDatabase(name)
	if err := db.Init(filePath); err != nil {
		return nil, err
	}

	return db, nil
}
