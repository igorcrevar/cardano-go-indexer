package core

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	ouroboros "github.com/blinklabs-io/gouroboros"
	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/protocol/chainsync"
	"github.com/blinklabs-io/gouroboros/protocol/common"
	"github.com/hashicorp/go-hclog"
)

const (
	ProtocolTCP  = "tcp"
	ProtocolUnix = "unix"
)

type GetTxsFunc func() ([]ledger.Transaction, error)

type BlockSyncer interface {
	Sync() error
	Close() error
}

type BlockSyncerHandler interface {
	RollBackwardFunc(point common.Point, tip chainsync.Tip) error
	RollForwardFunc(blockHeader *BlockHeader, getTxsFunc GetTxsFunc, tip chainsync.Tip) error
	SyncBlockPoint() (BlockPoint, error)
	NextBlockNumber() uint64
}

type BlockSyncerConfig struct {
	NetworkMagic   uint32        `json:"networkMagic"`
	NodeAddress    string        `json:"nodeAddress"`
	RestartOnError bool          `json:"restartOnError"`
	RestartDelay   time.Duration `json:"restartDelay"`
}

func (bsc BlockSyncerConfig) Protocol() string {
	if strings.HasPrefix(bsc.NodeAddress, "/") {
		return ProtocolUnix
	}

	return ProtocolTCP
}

type BlockSyncerImpl struct {
	connection   *ouroboros.Connection
	blockHandler BlockSyncerHandler
	config       *BlockSyncerConfig
	logger       hclog.Logger
}

var _ BlockSyncer = (*BlockSyncerImpl)(nil)

func NewBlockSyncer(config *BlockSyncerConfig, blockHandler BlockSyncerHandler, logger hclog.Logger) *BlockSyncerImpl {
	return &BlockSyncerImpl{
		blockHandler: blockHandler,
		config:       config,
		logger:       logger,
	}
}

func (bs *BlockSyncerImpl) Sync() error {
	blockPoint, err := bs.blockHandler.SyncBlockPoint()
	if err != nil {
		return err
	}

	bs.logger.Debug("Start syncing requested", "networkMagic", bs.config.NetworkMagic, "address", bs.config.NodeAddress, "slot", blockPoint.BlockSlot, "hash", hex.EncodeToString(blockPoint.BlockHash))

	if bs.connection != nil {
		bs.connection.Close() // close previous connection
	}

	// create connection
	connection, err := ouroboros.NewConnection(
		ouroboros.WithNetworkMagic(bs.config.NetworkMagic),
		ouroboros.WithNodeToNode(true),
		ouroboros.WithKeepAlive(true),
		ouroboros.WithChainSyncConfig(chainsync.NewConfig(
			chainsync.WithRollBackwardFunc(bs.blockHandler.RollBackwardFunc),
			chainsync.WithRollForwardFunc(bs.rollForwardCallback),
		)),
	)
	if err != nil {
		return err
	}

	bs.connection = connection

	// dial node -> connect to node
	if err := bs.connection.Dial(bs.config.Protocol(), bs.config.NodeAddress); err != nil {
		return err
	}

	var point common.Point

	if len(blockPoint.BlockHash) == 0 {
		point = common.NewPointOrigin() // from genesis
	} else {
		point = common.NewPoint(blockPoint.BlockSlot, blockPoint.BlockHash)
	}

	bs.logger.Debug("Syncing started", "networkMagic", bs.config.NetworkMagic, "address", bs.config.NodeAddress, "slot", blockPoint.BlockSlot, "hash", hex.EncodeToString(blockPoint.BlockHash))

	// start syncing
	if err := bs.connection.ChainSync().Client.Sync([]common.Point{point}); err != nil {
		return err
	}

	// in separated routine wait for async errors
	go bs.errorHandler()

	return nil
}

func (bs *BlockSyncerImpl) Close() error {
	if bs.connection == nil {
		return nil
	}

	return bs.connection.Close()
}

func (bs *BlockSyncerImpl) getBlock(slot uint64, hash []byte) (ledger.Block, error) {
	bs.logger.Debug("Get full block", "slot", slot, "hash", hex.EncodeToString(hash), "connected", bs.connection != nil)
	if bs.connection == nil {
		return nil, errors.New("no connection")
	}

	return bs.connection.BlockFetch().Client.GetBlock(common.NewPoint(slot, hash))
}

func (bs *BlockSyncerImpl) rollForwardCallback(blockType uint, blockInfo interface{}, tip chainsync.Tip) error {
	blockHeader, err := GetBlockHeaderFromBlockInfo(blockType, blockInfo, bs.blockHandler.NextBlockNumber())
	if err != nil {
		return errors.Join(errBlockIndexerFatal, err)
	}

	bs.logger.Debug("Roll forward", "number", blockHeader.BlockNumber,
		"hash", hex.EncodeToString(blockHeader.BlockHash), "slot", tip.Point.Slot, "hash", hex.EncodeToString(tip.Point.Hash))

	return bs.blockHandler.RollForwardFunc(blockHeader, func() ([]ledger.Transaction, error) {
		block, err := bs.getBlock(blockHeader.BlockSlot, blockHeader.BlockHash)
		if err != nil {
			return nil, err
		}

		return block.Transactions(), nil
	}, tip)
}

func (bs *BlockSyncerImpl) errorHandler() {
	if bs.connection == nil {
		return
	}

	err, ok := <-bs.connection.ErrorChan()
	if !ok {
		return
	}

	// retry syncing again if not fatal
	if !errors.Is(err, errBlockIndexerFatal) {
		bs.logger.Warn("Error happened", "err", err)

		time.Sleep(bs.config.RestartDelay)
		if err := bs.Sync(); err != nil {
			bs.logger.Warn("Error happened while trying to restart syncer", "err", err)
		}
	} else {
		bs.logger.Error("Fatal error happened", "err", err)
	}
}
