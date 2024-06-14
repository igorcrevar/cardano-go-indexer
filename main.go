package main

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/igorcrevar/cardano-go-indexer/core"
	"github.com/igorcrevar/cardano-go-indexer/db"

	"github.com/hashicorp/go-hclog"
)

func main() {
	networkMagic := uint32(42)
	address := "localhost:3000"   // "/tmp/cardano-133064331/node-spo1/node.sock"
	startBlockHash := []byte(nil) // from genesis
	startSlot := uint64(0)
	startBlockNum := uint64(math.MaxUint64)
	addressesOfInterest := []string{}

	// for test net
	address = "preprod-node.play.dev.cardano.org:3001"
	networkMagic = 1

	// for main net
	address = "backbone.cardano-mainnet.iohk.io:3001"
	networkMagic = uint32(764824073)

	startBlockHash, _ = hex.DecodeString("5d9435abf2a829142aaae08720afa05980efaa6ad58e47ebd4cffadc2f3c45d8")
	startSlot = uint64(76592549)
	startBlockNum = 7999980
	addressesOfInterest = []string{
		"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd",
	}

	logger, err := core.NewLogger(core.LoggerConfig{
		LogLevel:      hclog.Debug,
		JSONLogFormat: false,
		AppendFile:    true,
		LogFilePath:   "logs/cardano_indexer.log",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dbs, err := db.NewDatabaseInit("", "burek.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Open database failed", "err", err)
		os.Exit(1)
	}

	defer dbs.Close()

	confirmedBlockHandler := func(confirmedBlock *core.CardanoBlock, txs []*core.Tx) error {
		logger.Info("Confirmed block",
			"hash", hex.EncodeToString(confirmedBlock.Hash[:]), "slot", confirmedBlock.Slot,
			"allTxs", len(confirmedBlock.Txs), "ourTxs", len(txs))

		lastBlocks, err := dbs.GetLatestConfirmedBlocks(5)
		if err != nil {
			return err
		}

		lastBlocksInfo := make([]string, len(lastBlocks))
		for i, x := range lastBlocks {
			lastBlocksInfo[i] = fmt.Sprintf("(%d-%s)", x.Slot, hex.EncodeToString(x.Hash[:]))
		}

		logger.Debug("Last n blocks", "info", strings.Join(lastBlocksInfo, ", "))

		unprocessedTxs, err := dbs.GetUnprocessedConfirmedTxs(0)
		if err != nil {
			return err
		}

		for _, tx := range unprocessedTxs {
			logger.Info("Tx has been processed", "tx", tx)
		}

		return dbs.MarkConfirmedTxsProcessed(unprocessedTxs)
	}

	indexerConfig := &core.BlockIndexerConfig{
		StartingBlockPoint: &core.BlockPoint{
			BlockSlot:   startSlot,
			BlockHash:   core.Hash(startBlockHash),
			BlockNumber: startBlockNum,
		},
		AddressCheck:            core.AddressCheckAll,
		ConfirmationBlockCount:  10,
		AddressesOfInterest:     addressesOfInterest,
		SoftDeleteUtxo:          false,
		KeepAllTxOutputsInDB:    false,
		KeepAllTxsHashesInBlock: false,
	}
	syncerConfig := &core.BlockSyncerConfig{
		NetworkMagic:   networkMagic,
		NodeAddress:    address,
		RestartOnError: true,
		RestartDelay:   time.Second * 2,
		KeepAlive:      true,
	}

	indexer := core.NewBlockIndexer(indexerConfig, confirmedBlockHandler, dbs, logger.Named("block_indexer"))

	syncer := core.NewBlockSyncer(syncerConfig, indexer, logger.Named("block_syncer"))
	defer syncer.Close()

	err = syncer.Sync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		logger.Error("Start syncing failed", "err", err)
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	// Notify the signalChannel when the interrupt signal is received (Ctrl+C)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
	case <-syncer.ErrorCh():
	}
}
