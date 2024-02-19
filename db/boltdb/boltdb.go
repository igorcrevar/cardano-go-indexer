package boltdb

import (
	"encoding/json"
	"fmt"
	"igorcrevar/cardano-go-syncer/core"

	"github.com/boltdb/bolt"
)

type BoltDatabase struct {
	db *bolt.DB
}

var (
	txOutputsBucket         = []byte("TXOuts")
	latestBlockPointBucket  = []byte("LatestBlockPoint")
	processedBlocksBucket   = []byte("ProcessedBlocks")
	unprocessedBlocksBucket = []byte("UnprocessedBlocks")

	defaultKey = []byte("default")
)

var _ core.Database = (*BoltDatabase)(nil)

func (bd *BoltDatabase) Init(filePath string) error {
	db, err := bolt.Open(filePath, 0600, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %v", err)
	}

	bd.db = db

	return db.Update(func(tx *bolt.Tx) error {
		for _, bn := range [][]byte{txOutputsBucket, latestBlockPointBucket, processedBlocksBucket, unprocessedBlocksBucket} {
			_, err := tx.CreateBucketIfNotExists(bn)
			if err != nil {
				return fmt.Errorf("could not bucket: %s, err: %v", string(bn), err)
			}
		}

		return nil
	})
}

func (bd *BoltDatabase) Close() error {
	return bd.db.Close()
}

func (bd *BoltDatabase) GetLatestBlockPoint() (*core.BlockPoint, error) {
	var result *core.BlockPoint

	if err := bd.db.View(func(tx *bolt.Tx) error {
		if data := tx.Bucket(latestBlockPointBucket).Get(defaultKey); len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BoltDatabase) GetTxOutput(txInput core.TxInput) (*core.TxOutput, error) {
	var result *core.TxOutput

	if err := bd.db.View(func(tx *bolt.Tx) error {
		if data := tx.Bucket(txOutputsBucket).Get(txInput.Key()); len(data) > 0 {
			return json.Unmarshal(data, &result)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BoltDatabase) MarkConfirmedBlockProcessed(block *core.FullBlock) error {
	return bd.db.Update(func(tx *bolt.Tx) error {
		// move block from one bucket to the other
		data := tx.Bucket(unprocessedBlocksBucket).Get(block.Key())
		if len(data) == 0 {
			return fmt.Errorf("unprocessed block does not exist: %v", block.Key())
		}

		if err := tx.Bucket(unprocessedBlocksBucket).Delete(block.Key()); err != nil {
			return fmt.Errorf("could not set remove from unprocessed blocks: %v", err)
		}

		if err := tx.Bucket(processedBlocksBucket).Put(block.Key(), data); err != nil {
			return fmt.Errorf("could not move to processed blocks: %v", err)
		}

		return nil
	})
}

func (bd *BoltDatabase) GetUnprocessedConfirmedBlocks() ([]*core.FullBlock, error) {
	var result []*core.FullBlock

	if err := bd.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(unprocessedBlocksBucket).ForEach(func(k, v []byte) error {
			var block *core.FullBlock

			if err := json.Unmarshal(v, &block); err != nil {
				return err
			}

			result = append(result, block)

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (bd *BoltDatabase) OpenTx() core.DbTransactionWriter {
	return &BoltDbTransactionWriter{
		db: bd.db,
	}
}
