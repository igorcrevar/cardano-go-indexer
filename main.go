package main

import (
	"encoding/hex"
	"fmt"
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
	// for test net
	address := "preprod-node.play.dev.cardano.org:3001"
	networkMagic := uint32(1)
	startBlockHash, _ := hex.DecodeString("4b7f0ff899395dade775c7eb2bc8e16fb9824d9091266c7d4c9c55ac143ae6c8")
	startSlot := uint64(74683550)
	addressesOfInterest := []string{
		"addr_test1wr64gtafm8rpkndue4ck2nx95u4flhwf643l2qmg9emjajg2ww0nj",
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
			BlockSlot: startSlot,
			BlockHash: core.Hash(startBlockHash),
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
