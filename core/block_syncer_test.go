package core

import (
	"encoding/hex"
	"testing"

	ouroboros "github.com/blinklabs-io/gouroboros"
	"github.com/blinklabs-io/gouroboros/protocol/chainsync"
	"github.com/blinklabs-io/gouroboros/protocol/common"
	"github.com/hashicorp/go-hclog"
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

	if syncer := NewBlockSyncer(logger); syncer == nil {
		t.Fatalf("got NewBlockSyncer(uninitializedLogger) = %v; want = &BlockSyncer", syncer)
	}
}

func TestNewBlockSyncerWithInitializedLogger(t *testing.T) {
	logger := hclog.Default()

	if syncer := NewBlockSyncer(logger); syncer == nil {
		t.Fatalf("got NewBlockSyncer(initializedLogger) = %v; want = &BlockSyncer", syncer)
	}
}

func TestSyncWrongMagic(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	if err := syncer.Sync(71, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler); err == nil {
		t.Fatalf("got syncer.Sync(71, NodeAddress, ExistingPointSlot, existingPointHash, mockSyncerBlockHandler) = %v; want = handshake: refused", err)
	}
}

func TestSyncWrongNodeAddress(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	if err := syncer.Sync(NetworkMagic, "test", ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler); err == nil {
		t.Fatalf(`got syncer.Sync(NetworkMagic, "test", ExistingPointSlot, existingPointHash, mockSyncerBlockHandler) = %v; want = dial tcp`, err)
	}
}

func TestSyncWrongUnixNodeAddress(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	if err := syncer.Sync(NetworkMagic, "/"+NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler); err == nil {
		t.Fatalf(`got syncer.Sync(NetworkMagic, "/"+NodeAddress, ExistingPointSlot, existingPointHash, mockSyncerBlockHandler) = %v; want = dial unix`, err)
	}
}

func TestSyncNonExistingSlot(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	if err := syncer.Sync(NetworkMagic, NodeAddress, NonExistingPointSlot, existingPointHash, &mockSyncerBlockHandler); err == nil {
		t.Fatalf("got syncer.Sync(NetworkMagic, NodeAddress, NonExistingPointSlot, existingPointHash, mockSyncerBlockHandler) = %v; want = chain intersection not found", err)
	}
}

func TestSyncNonExistingHash(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	nonExistingPointHash, _ := hex.DecodeString(NonExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	if err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, nonExistingPointHash, &mockSyncerBlockHandler); err == nil {
		t.Fatalf("got syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, nonExistingPointHash, mockSyncerBlockHandler) = %v; want = chain intersection not found", err)
	}
}

func TestSyncEmptyHash(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	var emptyHash []byte

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	if err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, emptyHash, &mockSyncerBlockHandler); err != nil {
		t.Fatalf("got syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, emptyHash, mockSyncerBlockHandler) = %v; want = nil", err)
	}
}

func TestSync(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	if err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler); err != nil {
		t.Fatalf("got syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, mockSyncerBlockHandler) = %v; want = nil", err)
	}
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
	if err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler); err != nil {
		t.Fatalf("got syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, mockSyncerBlockHandler) = %v; want = nil", err)
	}
}

func TestCloseWithConnectionNil(t *testing.T) {
	syncer := NewBlockSyncer(nil)

	if err := syncer.Close(); err != nil {
		t.Fatalf("got syncer.Close() = %v; want = nil", err)
	}
}

func TestCloseWithConnectionNotNil(t *testing.T) {
	syncer := NewBlockSyncer(nil)
	connection, _ := ouroboros.NewConnection(
		ouroboros.WithNetworkMagic(NetworkMagic),
		ouroboros.WithNodeToNode(true),
		ouroboros.WithKeepAlive(true),
	)
	syncer.connection = connection

	if err := syncer.Close(); err != nil {
		t.Fatalf("got syncer.Close() = %v; want = nil", err)
	}
}

func TestGetFullBlockWithConnectionNil(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())

	var slot uint64
	var hash []byte
	if _, err := syncer.GetFullBlock(slot, hash); err == nil {
		t.Fatalf("got syncer.GetFullBlock(slot, hash) = (_, %v); want = (_, no connection)", err)
	}
}

func TestGetFullBlockWithConnectionNotNilExisting(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingPointHash, _ := hex.DecodeString(ExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	// .GetFullBlock panics without full initialize from .Sync
	if err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingPointHash, &mockSyncerBlockHandler); err != nil {
		panic("failed to start sync")
	}

	if block, err := syncer.GetFullBlock(ExistingPointSlot, existingPointHash); block == nil || err != nil {
		t.Fatalf("got syncer.GetFullBlock(ExistingPointSlot, existingPointHash) = (%v, %v); want = (block, nil)", block, err)
	}
}

func TestGetFullBlockWithConnectionNotNilNotExisting(t *testing.T) {
	syncer := NewBlockSyncer(hclog.Default())
	existingHash, _ := hex.DecodeString(ExistingPointHashStr)
	nonExistingPointHash, _ := hex.DecodeString(NonExistingPointHashStr)

	mockSyncerBlockHandler := BlockSyncerHandlerMock{}
	defer syncer.Close()
	// .GetFullBlock panics without full initialize from .Sync
	if err := syncer.Sync(NetworkMagic, NodeAddress, ExistingPointSlot, existingHash, &mockSyncerBlockHandler); err != nil {
		panic("failed to start sync")
	}

	if block, err := syncer.GetFullBlock(NonExistingPointSlot, nonExistingPointHash); block != nil || err == nil {
		t.Fatalf("got syncer.GetFullBlock(NonExistingPointSlot, nonExistingPointHash) = (%v, %v); want = (nil, block(s) not found)", block, err)
	}
}
