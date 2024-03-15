package db

import (
	"strings"

	"github.com/igorcrevar/cardano-go-indexer/core"
	"github.com/igorcrevar/cardano-go-indexer/db/bbolt"
	"github.com/igorcrevar/cardano-go-indexer/db/leveldb"
)

func NewDatabase(name string) core.Database {
	switch strings.ToLower(name) {
	case "leveldb":
		return &leveldb.LevelDbDatabase{}
	default:
		return &bbolt.BBoltDatabase{}
	}
}

func NewDatabaseInit(name string, filePath string) (core.Database, error) {
	db := NewDatabase(name)
	if err := db.Init(filePath); err != nil {
		return nil, err
	}

	return db, nil
}
