package core

import (
	"errors"
	"strings"

	ouroboros "github.com/blinklabs-io/gouroboros"
	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/protocol/chainsync"
	"github.com/blinklabs-io/gouroboros/protocol/common"
)

const (
	ProtocolTCP  = "tcp"
	ProtocolUnix = "unix"
)

type BlockSyncer interface {
	Sync(networkMagic uint32, nodeAddress string, slot uint64, blockHash []byte, blockHandler BlockSyncerHandler) error
	GetFullBlock(slot uint64, hash []byte) (ledger.Block, error)
	Close() error
}

type BlockSyncerHandler interface {
	RollBackwardFunc(point common.Point, tip chainsync.Tip) error
	RollForwardFunc(blockType uint, blockInfo interface{}, tip chainsync.Tip) error
	ErrorHandler(err error)
}

type BlockSyncerImpl struct {
	connection *ouroboros.Connection
}

var _ BlockSyncer = (*BlockSyncerImpl)(nil)

func NewBlockSyncer() *BlockSyncerImpl {
	return &BlockSyncerImpl{}
}

func (bs *BlockSyncerImpl) Sync(networkMagic uint32, nodeAddress string, slot uint64, blockHash []byte, blockHandler BlockSyncerHandler) error {
	if bs.connection != nil {
		bs.connection.Close() // close previous connection
	}

	// create connection
	connection, err := ouroboros.NewConnection(
		ouroboros.WithNetworkMagic(networkMagic),
		ouroboros.WithNodeToNode(true),
		ouroboros.WithKeepAlive(true),
		ouroboros.WithChainSyncConfig(chainsync.NewConfig(
			chainsync.WithRollBackwardFunc(blockHandler.RollBackwardFunc),
			chainsync.WithRollForwardFunc(blockHandler.RollForwardFunc),
		)),
	)
	if err != nil {
		return err
	}

	bs.connection = connection

	proto := ProtocolTCP
	if strings.HasPrefix(nodeAddress, "/") {
		proto = ProtocolUnix
	}

	// dial node -> connect to node
	if err := bs.connection.Dial(proto, nodeAddress); err != nil {
		return err
	}

	var point common.Point

	if len(blockHash) == 0 {
		point = common.NewPointOrigin() // from genesis
	} else {
		point = common.NewPoint(slot, blockHash)
	}

	// start syncing
	if err := bs.connection.ChainSync().Client.Sync([]common.Point{point}); err != nil {
		return err
	}

	// in separated routine wait for async errors
	go func() {
		err, ok := <-bs.connection.ErrorChan()
		if !ok {
			return
		}

		blockHandler.ErrorHandler(err)
	}()

	return nil
}

func (bs *BlockSyncerImpl) Close() error {
	if bs.connection == nil {
		return nil
	}

	return bs.connection.Close()
}

func (bs *BlockSyncerImpl) GetFullBlock(slot uint64, hash []byte) (ledger.Block, error) {
	if bs.connection == nil {
		return nil, errors.New("no connection")
	}

	return bs.connection.BlockFetch().Client.GetBlock(common.NewPoint(slot, hash))
}
