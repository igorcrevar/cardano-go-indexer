package core

import (
	"encoding/hex"
	"testing"

	ouroboros "github.com/blinklabs-io/gouroboros"
	"github.com/blinklabs-io/gouroboros/protocol/chainsync"
	"github.com/blinklabs-io/gouroboros/protocol/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

type BlockSyncerHandlerMock struct{}

func (hMock *BlockSyncerHandlerMock) RollForwardFunc(blockType uint, blockInfo interface{}, tip chainsync.Tip) error {
	return nil
}
func (hMock *BlockSyncerHandlerMock) RollBackwardFunc(point common.Point, tip chainsync.Tip) error {
	return nil
}
func (hMock *BlockSyncerHandlerMock) ErrorHandler(err error) {
}

const (
	NodeAddress             = "preprod-node.play.dev.cardano.org:3001"
	NetworkMagic            = 1
	ExistingPointSlot       = uint64(2607239)
	ExistingPointHashStr    = "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d19"
	NonExistingPointSlot    = uint64(2607240)
	NonExistingPointHashStr = "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d20"
)

func TestNewBlockSyncerWithUninitializedLogger(t *testing.T) {
	var logger hclog.Logger

	syncer := NewBlockSyncer(logger)
	require.NotNil(t, syncer)
}

func TestNewBlockSyncerWithInitializedLogger(t *testing.T) {
	logger := hclog.Default()

	syncer := NewBlockSyncer(logger)
	require.NotNil(t, syncer)
}

func TestSyncWrongMagic(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(71, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.NotNil(t, err)
}

func TestSyncWrongNodeAddress(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(NetworkMagic, "test", ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.NotNil(t, err)
}

func TestSyncWrongUnixNodeAddress(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(NetworkMagic, "/"+NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.NotNil(t, err)
}

func TestSyncNonExistingSlot(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(NetworkMagic, NodeAddress, NonExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.NotNil(t, err)
}

func TestSyncNonExistingHash(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	nonExistingPointHash, _ := hex.DecodeString(NonExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, nonExistingPointHash, &mockSyncerBlockHandler)
	require.NotNil(t, err)
}

func TestSyncEmptyHash(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	var emptyHash []byte

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, emptyHash, &mockSyncerBlockHandler)
	require.Nil(t, err)
}

func TestSync(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.Nil(t, err)
}

func TestSyncWithExistingConnection(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	connection, _ := ouroboros.NewConnection(
		ouroboros.WithNetworkMagic(NetworkMagic),
		ouroboros.WithNodeToNode(true),
		ouroboros.WithKeepAlive(true),
	)
	syncer.connection = connection
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.Nil(t, err)
}

func TestCloseWithConnectionNil(t *testing.T) {
	syncer := NewBlockSyncer(nil)

	err := syncer.Close()
	require.Nil(t, err)
}

func TestCloseWithConnectionNotNil(t *testing.T) {
	syncer := NewBlockSyncer(nil)
	connection, _ := ouroboros.NewConnection(
		ouroboros.WithNetworkMagic(NetworkMagic),
		ouroboros.WithNodeToNode(true),
		ouroboros.WithKeepAlive(true),
	)
	syncer.connection = connection

	err := syncer.Close()
	require.Nil(t, err)
}

func TestGetFullBlockWithConnectionNil(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())

	var slot uint64
	var hash []byte

	_, err := syncer.GetFullBlock(slot, hash)
	require.NotNil(t, err)
}

func TestGetFullBlockWithConnectionNotNilExisting(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	// .GetFullBlock panics without full initialize from .Sync
	err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.Nil(t, err, "failed to start sync")

	block, err := syncer.GetFullBlock(ExistingPointSlot, existingPointHash)
	require.NotNil(t, block)
	require.Nil(t, err)
}

func TestGetFullBlockWithConnectionNotNilNotExisting(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)
	nonExistingPointHash, _ := hex.DecodeString(NonExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	// .GetFullBlock panics without full initialize from .Sync
	err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler)
	require.Nil(t, err, "failed to start sync")

	block, err := syncer.GetFullBlock(NonExistingPointSlot, nonExistingPointHash)
	require.Nil(t, block)
	require.NotNil(t, err)
}
