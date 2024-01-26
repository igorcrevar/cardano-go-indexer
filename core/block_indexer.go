package core

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/protocol/chainsync"
	"github.com/blinklabs-io/gouroboros/protocol/common"
)

var (
	errBlockIndexerFatal = errors.New("block indexer fatal error")
)

type BlockIndexerConfig struct {
	NetworkMagic uint32 `json:"networkMagic"`
	NodeAddress  string `json:"nodeAddress"`

	StartingBlockPoint *BlockPoint `json:"startingBlockPoint"`

	// how many children blocks is needed for some block to be considered final
	ConfirmationBlockCount uint `json:"confirmationBlockCount"`

	AddressesOfInterest []string `json:"addressesOfInterest"`
}

type NewConfirmedBlockHandler func(*FullBlock) error

type BlockIndexer struct {
	blockSyncer BlockSyncer
	config      *BlockIndexerConfig

	// latest confirmed and saved block point
	latestBlockPoint *BlockPoint

	newConfirmedBlockHandler NewConfirmedBlockHandler
	unconfirmedBlocks        []*BlockHeader

	db                  BlockIndexerDb
	addressesOfInterest map[string]bool
}

var _ BlockSyncerHandler = (*BlockIndexer)(nil)

func NewBlockIndexer(config *BlockIndexerConfig, blockSyncer BlockSyncer, newConfirmedBlockHandler NewConfirmedBlockHandler, db BlockIndexerDb) *BlockIndexer {
	addressesOfInterest := make(map[string]bool, len(config.AddressesOfInterest))
	for _, x := range config.AddressesOfInterest {
		addressesOfInterest[x] = true
	}

	return &BlockIndexer{
		blockSyncer: blockSyncer,
		config:      config,

		latestBlockPoint: nil,

		newConfirmedBlockHandler: newConfirmedBlockHandler,
		unconfirmedBlocks:        nil,

		db:                  db,
		addressesOfInterest: addressesOfInterest,
	}
}

func (bi *BlockIndexer) RollBackwardFunc(point common.Point, tip chainsync.Tip) error {
	// linear is ok, there will be smaller number of unconfirmed blocks in memory
	for i := len(bi.unconfirmedBlocks) - 1; i >= 0; i-- {
		unc := bi.unconfirmedBlocks[i]
		if unc.BlockSlot == point.Slot && bytes.Equal(unc.BlockHash, point.Hash) {
			bi.unconfirmedBlocks = bi.unconfirmedBlocks[:i+1]

			return nil
		}
	}

	if bi.latestBlockPoint.BlockSlot == point.Slot && bytes.Equal(bi.latestBlockPoint.BlockHash, point.Hash) {
		// everything is ok -> we are reverting to the latest confirmed block
		return nil
	}

	// we have confirmed some block that should not be confirmed!!!! TODO: what to do in this case?
	return errors.Join(errBlockIndexerFatal, fmt.Errorf("roll backward, block not found = (%d, %s)", point.Slot, hex.EncodeToString(point.Hash)))
}

func (bi *BlockIndexer) RollForwardFunc(blockType uint, blockInfo interface{}, tip chainsync.Tip) error {
	nextBlockNumber := bi.latestBlockPoint.BlockNumber + 1
	if len(bi.unconfirmedBlocks) > 0 {
		nextBlockNumber = bi.unconfirmedBlocks[len(bi.unconfirmedBlocks)-1].BlockNumber + 1
	}

	blockHeader, err := GetBlockHeaderFromBlockInfo(blockType, blockInfo, nextBlockNumber)
	if err != nil {
		return errors.Join(errBlockIndexerFatal, err)
	}

	fmt.Printf("roll forward: number = %d, hash = %s, tip block = %d, tip point = (%d, %s)\n",
		blockHeader.BlockNumber, hex.EncodeToString(blockHeader.BlockHash), tip.BlockNumber, tip.Point.Slot, hex.EncodeToString(tip.Point.Hash))

	isFirstBlockConfirmed := uint(len(bi.unconfirmedBlocks)) >= bi.config.ConfirmationBlockCount

	var confirmedBlockHeader *BlockHeader
	if isFirstBlockConfirmed {
		confirmedBlockHeader = bi.unconfirmedBlocks[0]
	}

	fullBlock, latestBlockPoint, err := bi.processNewConfirmedBlock(confirmedBlockHeader)
	if err != nil {
		return err
	}

	if isFirstBlockConfirmed {
		// update latest block point in memory if we have confirmed block
		bi.latestBlockPoint = latestBlockPoint
		// remove first block from unconfirmed list. copy whole list because we do not want memory leak
		bi.unconfirmedBlocks = append([]*BlockHeader(nil), bi.unconfirmedBlocks[1:]...)
	}

	bi.unconfirmedBlocks = append(bi.unconfirmedBlocks, blockHeader)

	// notify listener if needed
	if fullBlock != nil {
		bi.newConfirmedBlockHandler(fullBlock)
	}

	return nil
}

func (bi *BlockIndexer) ErrorHandler(err error) {
	fmt.Printf("error retrieved: %v\n", err)

	// retry syncing again if not fatal
	if !errors.Is(err, errBlockIndexerFatal) {
		if err := bi.StartSyncing(); err != nil {
			fmt.Printf("failed to retry syncing: %v\n", err)
		}
	}
}

func (bi *BlockIndexer) StartSyncing() error {
	if bi.latestBlockPoint == nil {
		// read from database
		latestBlockPoint, err := bi.db.GetLatestBlockPoint()
		if err != nil {
			return err
		}

		bi.latestBlockPoint = latestBlockPoint
		// if there is nothing in database read from default config
		if bi.latestBlockPoint == nil {
			bi.latestBlockPoint = bi.config.StartingBlockPoint
		}

		if bi.latestBlockPoint == nil {
			bi.latestBlockPoint = &BlockPoint{
				BlockSlot:   0,
				BlockNumber: math.MaxUint64,
				BlockHash:   nil,
			}
		}
	}

	return bi.blockSyncer.Sync(bi.config.NetworkMagic, bi.config.NodeAddress, bi.latestBlockPoint.BlockSlot, bi.latestBlockPoint.BlockHash, bi)
}

func (bi *BlockIndexer) Close() error {
	return bi.blockSyncer.Close()
}

func (bi *BlockIndexer) processNewConfirmedBlock(confirmedBlockHeader *BlockHeader) (*FullBlock, *BlockPoint, error) {
	if confirmedBlockHeader == nil {
		return nil, bi.latestBlockPoint, nil
	}

	block, err := bi.blockSyncer.GetFullBlock(confirmedBlockHeader.BlockSlot, confirmedBlockHeader.BlockHash)
	if err != nil {
		return nil, nil, err
	}

	var fullBlock *FullBlock = nil

	// open database tx
	dbTx := bi.db.OpenTx()

	// get all transactions of interesy from block
	// if there is none, we do not need to process this block further
	blockTransactions, err := bi.getTxsOfInterest(block.Transactions())
	if err != nil {
		return nil, nil, err
	} else if len(blockTransactions) > 0 {
		fullBlock = NewFullBlock(confirmedBlockHeader, NewTransactions(blockTransactions))

		dbTx.AddConfirmedBlock(fullBlock)
		bi.addTxOutputs(dbTx, fullBlock)
	}

	latestBlockPoint := &BlockPoint{
		BlockSlot:   confirmedBlockHeader.BlockSlot,
		BlockHash:   confirmedBlockHeader.BlockHash,
		BlockNumber: confirmedBlockHeader.BlockNumber,
	}
	// update database -> execute db transaction
	dbTx.SetLatestBlockPoint(bi.latestBlockPoint)
	if err := dbTx.Execute(); err != nil {
		return nil, nil, err
	}

	return fullBlock, latestBlockPoint, nil
}

func (bi *BlockIndexer) getTxsOfInterest(txs []ledger.Transaction) (result []ledger.Transaction, err error) {
	if len(bi.addressesOfInterest) == 0 {
		return txs, nil
	}

	for _, tx := range txs {
		if bi.isTxOutputOfInterest(tx) {
			result = append(result, tx)
		} else {
			txIsGood, err := bi.isTxInputOfInterest(tx)
			if err != nil {
				return nil, err
			} else if txIsGood {
				result = append(result, tx)
			}
		}
	}

	return result, nil
}

func (bi *BlockIndexer) isTxOutputOfInterest(tx ledger.Transaction) bool {
	for _, out := range tx.Outputs() {
		address := out.Address().String()
		if bi.addressesOfInterest[address] {
			return true
		}
	}

	return false
}

func (bi *BlockIndexer) isTxInputOfInterest(tx ledger.Transaction) (bool, error) {
	for _, inp := range tx.Inputs() {
		txOutput, err := bi.db.GetTxOutput(TxInput{
			Hash:  inp.Id().String(),
			Index: inp.Index(),
		})
		if err != nil {
			return false, err
		} else if txOutput != nil && bi.addressesOfInterest[txOutput.Address] {
			return true, nil
		}
	}

	return false, nil
}

func (bi *BlockIndexer) addTxOutputs(dbTx DbTransactionWriter, block *FullBlock) {
	for _, tx := range block.Txs {
		for ind, txOut := range tx.Outputs {
			if bi.addressesOfInterest[txOut.Address] {
				// add tx output to database
				dbTx.AddTxOutput(TxInput{
					Hash:  tx.Hash,
					Index: uint32(ind),
				}, txOut)
			}
		}
	}
}
