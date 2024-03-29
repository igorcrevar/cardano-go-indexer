package leveldb

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/igorcrevar/cardano-go-indexer/core"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LevelDbDatabase struct {
	db *leveldb.DB
}

var (
	txOutputsBucket        = []byte("P1_")
	latestBlockPointBucket = []byte("P2_")
	processedTxsBucket     = []byte("P3_")
	unprocessedTxsBucket   = []byte("P4_")
	confirmedBlocks        = []byte("P5_")
)

var _ core.Database = (*LevelDbDatabase)(nil)

func (lvldb *LevelDbDatabase) Init(filePath string) error {
	db, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		return fmt.Errorf("could not open db: %v", err)
	}

	lvldb.db = db

	return nil
}

func (bd *LevelDbDatabase) Close() error {
	return bd.db.Close()
}

func (lvldb *LevelDbDatabase) GetLatestBlockPoint() (*core.BlockPoint, error) {
	var result *core.BlockPoint

	bytes, err := lvldb.db.Get(latestBlockPointBucket, nil)
	if err != nil {
		return nil, processNotFoundErr(err)
	}

	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (lvldb *LevelDbDatabase) GetTxOutput(txInput core.TxInput) (result core.TxOutput, err error) {
	bytes, err := lvldb.db.Get(bucketKey(txOutputsBucket, txInput.Key()), nil)
	if err != nil {
		return result, processNotFoundErr(err)
	}

	err = json.Unmarshal(bytes, &result)

	return result, err
}

func (lvldb *LevelDbDatabase) MarkConfirmedTxsProcessed(txs []*core.Tx) error {
	batch := new(leveldb.Batch)

	for _, tx := range txs {
		bytes, err := json.Marshal(tx)
		if err != nil {
			return fmt.Errorf("could not marshal tx: %v", err)
		}

		batch.Put(bucketKey(processedTxsBucket, tx.Key()), bytes)
		batch.Delete(bucketKey(unprocessedTxsBucket, tx.Key()))
	}

	return lvldb.db.Write(batch, &opt.WriteOptions{
		NoWriteMerge: false,
		Sync:         true,
	})
}

func (lvldb *LevelDbDatabase) GetUnprocessedConfirmedTxs(maxCnt int) ([]*core.Tx, error) {
	var result []*core.Tx

	iter := lvldb.db.NewIterator(util.BytesPrefix(unprocessedTxsBucket), nil)
	defer iter.Release()

	for iter.Next() {
		var tx *core.Tx

		if err := json.Unmarshal(iter.Value(), &tx); err != nil {
			return nil, err
		}

		result = append(result, tx)
		if maxCnt > 0 && len(result) == maxCnt {
			break
		}
	}

	return result, iter.Error()
}

func (lvldb *LevelDbDatabase) GetLatestConfirmedBlocks(maxCnt int) ([]*core.CardanoBlock, error) {
	var result []*core.CardanoBlock

	iter := lvldb.db.NewIterator(util.BytesPrefix(confirmedBlocks), nil)
	defer iter.Release()

	for ok := iter.Last(); ok; ok = iter.Prev() {
		var block *core.CardanoBlock

		if err := json.Unmarshal(iter.Value(), &block); err != nil {
			return nil, err
		}

		result = append(result, block)
		if maxCnt > 0 && len(result) == maxCnt {
			break
		}
	}

	return result, iter.Error()
}

func (lvldb *LevelDbDatabase) GetConfirmedBlocksFrom(slotNumber uint64, maxCnt int) ([]*core.CardanoBlock, error) {
	var result []*core.CardanoBlock

	iter := lvldb.db.NewIterator(util.BytesPrefix(confirmedBlocks), nil)
	defer iter.Release()

	for ok := iter.Seek(bucketKey(confirmedBlocks, core.SlotNumberToKey(slotNumber))); ok; ok = iter.Next() {
		var block *core.CardanoBlock

		if err := json.Unmarshal(iter.Value(), &block); err != nil {
			return nil, err
		}

		result = append(result, block)
		if maxCnt > 0 && len(result) == maxCnt {
			break
		}
	}

	return result, iter.Error()
}

func (lvldb *LevelDbDatabase) OpenTx() core.DbTransactionWriter {
	return NewLevelDbTransactionWriter(lvldb.db)
}

func bucketKey(bucket []byte, key []byte) []byte {
	const separator = "_#_"

	outputKey := make([]byte, len(bucket)+len(separator)+len(key))
	copy(outputKey, bucket)
	copy(outputKey[len(bucket):], []byte(separator))
	copy(outputKey[len(bucket)+len(separator):], key)

	return outputKey
}

func processNotFoundErr(err error) error {
	if errors.Is(err, leveldb.ErrNotFound) {
		return nil
	}

	return err
}
