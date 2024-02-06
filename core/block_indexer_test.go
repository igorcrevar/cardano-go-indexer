package core

import (
	"math"
	"testing"

	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/protocol/chainsync"
	"github.com/blinklabs-io/gouroboros/protocol/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBlockSyncer
type MockBlockSyncer struct {
	mock.Mock
}

func (m *MockBlockSyncer) Sync(networkMagic uint32, nodeAddress string, startingSlot uint64, startingHash []byte, handler BlockSyncerHandler) error {
	args := m.Called(networkMagic, nodeAddress, startingSlot, startingHash, handler)
	return args.Error(0)
}

func (m *MockBlockSyncer) GetFullBlock(slot uint64, hash []byte) (ledger.Block, error) {
	args := m.Called(slot, hash)
	return args.Get(0).(ledger.Block), args.Error(1)
}

func (m *MockBlockSyncer) Close() error {
	return nil
}

// MockBlockIndexerDb
type MockBlockIndexerDb struct {
	mock.Mock
}

func (m *MockBlockIndexerDb) GetLatestBlockPoint() (*BlockPoint, error) {
	args := m.Called()
	return args.Get(0).(*BlockPoint), args.Error(1)
}

func (m *MockBlockIndexerDb) OpenTx() DbTransactionWriter {
	m.Called()
	return new(MockDbTransactionWriter)
}

func (m *MockBlockIndexerDb) GetTxOutput(input TxInput) (*TxOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*TxOutput), args.Error(1)
}

// MockDbTransactionWriter
type MockDbTransactionWriter struct {
	mock.Mock
}

func NewMockDbTransactionWriter() *MockDbTransactionWriter {
	mockDbTxWriter := new(MockDbTransactionWriter)
	return mockDbTxWriter
}

func (m *MockDbTransactionWriter) AddConfirmedBlock(block *FullBlock) DbTransactionWriter {
	m.Called(block)
	return m
}

func (m *MockDbTransactionWriter) SetLatestBlockPoint(point *BlockPoint) DbTransactionWriter {
	m.Called(point)
	return m
}

func (m *MockDbTransactionWriter) AddTxOutput(input TxInput, output *TxOutput) DbTransactionWriter {
	// args := m.Called(input, output)
	// return args.Get(0).(*MockDbTransactionWriter)
	return m
}

func (m *MockDbTransactionWriter) RemoveTxOutputs(txInputs []*TxInput) DbTransactionWriter {
	m.Called(txInputs)
	return m
}

func (m *MockDbTransactionWriter) Execute() error {
	return nil
}

func TestNewBlockIndexer(t *testing.T) {
	mockBlockSyncer := new(MockBlockSyncer)
	mockDb := new(MockBlockIndexerDb)
	mockLogger := hclog.NewNullLogger()
	config := GetDummyConfig()

	blockIndexer := NewBlockIndexer(config, nil, mockBlockSyncer, mockDb, mockLogger)

	// Assertions
	assert.NotNil(t, blockIndexer)
	assert.Equal(t, config, blockIndexer.config)
	assert.NotNil(t, blockIndexer.addressesOfInterest)
	assert.Equal(t, len(blockIndexer.addressesOfInterest), 2)
	assert.NotNil(t, blockIndexer.logger)
}

func TestRollBackwardFunc_Unconfirmed(t *testing.T) {
	config := GetDummyConfig()

	blockIndexer := NewBlockIndexer(config, nil, nil, nil, nil)

	blockIndexer.latestBlockPoint = GetDummyBlockPoint(10, "01", 1)
	blockIndexer.unconfirmedBlocks = []*BlockHeader{
		GetDummyBlockHeader(20, "02", 2),
		GetDummyBlockHeader(30, "03", 3),
		GetDummyBlockHeader(40, "04", 4),
		GetDummyBlockHeader(50, "05", 5),
		GetDummyBlockHeader(60, "06", 6),
	}

	point := GetDummyBlockPoint(30, "03", 3)
	cardano_point := common.NewPoint(point.BlockSlot, point.BlockHash)

	err := blockIndexer.RollBackwardFunc(cardano_point, chainsync.Tip{})

	assert.NoError(t, err)
	assert.EqualValues(t, blockIndexer.unconfirmedBlocks, []*BlockHeader{
		GetDummyBlockHeader(20, "02", 2),
		GetDummyBlockHeader(30, "03", 3),
	})
}

func TestRollBackwardFunc_LatestBlock(t *testing.T) {
	config := GetDummyConfig()

	blockIndexer := NewBlockIndexer(config, nil, nil, nil, nil)

	blockIndexer.latestBlockPoint = GetDummyBlockPoint(10, "01", 1)
	blockIndexer.unconfirmedBlocks = []*BlockHeader{
		GetDummyBlockHeader(20, "02", 2),
		GetDummyBlockHeader(30, "03", 3),
		GetDummyBlockHeader(40, "04", 4),
		GetDummyBlockHeader(50, "05", 5),
		GetDummyBlockHeader(60, "06", 6),
	}

	point := GetDummyBlockPoint(10, "01", 1)
	cardano_point := common.NewPoint(point.BlockSlot, point.BlockHash)

	err := blockIndexer.RollBackwardFunc(cardano_point, chainsync.Tip{})

	assert.NoError(t, err)
	assert.EqualValues(t, blockIndexer.unconfirmedBlocks, []*BlockHeader{
		GetDummyBlockHeader(20, "02", 2),
		GetDummyBlockHeader(30, "03", 3),
		GetDummyBlockHeader(40, "04", 4),
		GetDummyBlockHeader(50, "05", 5),
		GetDummyBlockHeader(60, "06", 6),
	})
}

// TODO get blockinfo and tip objects from cardano chain
func TestRollForwardFunc(t *testing.T) {
	config := GetDummyConfig()

	blockIndexer := NewBlockIndexer(config, nil, nil, nil, nil)

	_ = blockIndexer.RollForwardFunc(1, nil, chainsync.Tip{})

	// BlockNumber
	// BlockHash

	// Hash
	// BlockNumber
	// SlotNumber
	// Era

	// my BlockHeader
	// bi.unconfirmed blocks da ima my BlockHeader
}

func TestStartSyncing_FromBegining(t *testing.T) {
	mockBlockSyncer := new(MockBlockSyncer)
	mockDb := new(MockBlockIndexerDb)
	config := GetDummyConfig()

	blockIndexer := NewBlockIndexer(config, nil, mockBlockSyncer, mockDb, nil)

	mockDb.On("GetLatestBlockPoint").Return(nil, nil)
	mockBlockSyncer.On("Sync", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := blockIndexer.StartSyncing()

	blockPoint := GetDummyBlockPoint(0, "00", math.MaxUint64)
	blockPoint.BlockHash = nil

	assert.EqualValues(t, blockIndexer.latestBlockPoint, blockPoint)
	assert.NoError(t, err)
}

func TestStartSyncing_FromPoint(t *testing.T) {
	mockBlockSyncer := new(MockBlockSyncer)
	mockDb := new(MockBlockIndexerDb)
	config := GetDummyConfig()

	blockIndexer := NewBlockIndexer(config, nil, mockBlockSyncer, mockDb, nil)

	blockPoint := GetDummyBlockPoint(5, "05", 5)
	mockDb.On("GetLatestBlockPoint").Return(blockPoint, nil)
	mockBlockSyncer.On("Sync", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := blockIndexer.StartSyncing()

	assert.EqualValues(t, blockIndexer.latestBlockPoint, blockPoint)
	assert.NoError(t, err)
}

func TestProcessNewConfirmedBlock(t *testing.T) {
	assert.True(t, false)
}

func TestGetTxsOfInterest(t *testing.T) {
	assert.True(t, false)
	// TODO: Combine tx input and output interest test
}

func TestIsTxOutputOfInterest_True(t *testing.T) {
	config := GetDummyConfig()
	blockIndexer := NewBlockIndexer(config, nil, nil, nil, nil)
	tx := GetDummyShelleyTransaction_WithInterest()

	isTxOfInterest := blockIndexer.isTxOutputOfInterest(tx)

	assert.NotNil(t, blockIndexer.addressesOfInterest)
	assert.Equal(t, len(blockIndexer.addressesOfInterest), 2)

	assert.True(t, isTxOfInterest)
}

func TestIsTxOutputOfInterest_False(t *testing.T) {
	config := GetDummyConfig()
	blockIndexer := NewBlockIndexer(config, nil, nil, nil, nil)
	tx := GetDummyShelleyTransaction_WithoutInterest()

	isTxOfInterest := blockIndexer.isTxOutputOfInterest(tx)

	assert.NotNil(t, blockIndexer.addressesOfInterest)
	assert.Equal(t, len(blockIndexer.addressesOfInterest), 2)

	assert.False(t, isTxOfInterest)
}

func TestIsTxInputOfInterest_True(t *testing.T) {
	config := GetDummyConfig()
	blockIndexer := NewBlockIndexer(config, nil, nil, nil, nil)
	tx := GetDummyShelleyTransaction_WithInterest()

	isTxOfInterest, _ := blockIndexer.isTxInputOfInterest(tx)

	assert.NotNil(t, blockIndexer.addressesOfInterest)
	assert.Equal(t, len(blockIndexer.addressesOfInterest), 2)

	assert.True(t, isTxOfInterest)
}
func TestIsTxInputOfInterest_False(t *testing.T) {
	assert.True(t, false)
	// TODO:
}

func TestAddTxOutputsToDb(t *testing.T) {
	mockBlockSyncer := new(MockBlockSyncer)
	mockDb := new(MockBlockIndexerDb)
	mockLogger := hclog.NewNullLogger()
	config := GetDummyConfig()

	mockTxWritter := NewMockDbTransactionWriter()
	dummyBlock := GetDummyFullBlock(1, "01", 1)

	blockIndexer := NewBlockIndexer(config, nil, mockBlockSyncer, mockDb, mockLogger)

	blockIndexer.addTxOutputsToDb(mockTxWritter, dummyBlock.Txs)
	mockTxWritter.Execute()

	// Assertions
	// TODO:
}

func TestAddAllTxOutputsToDb(t *testing.T) {
	mockBlockSyncer := new(MockBlockSyncer)
	mockDb := new(MockBlockIndexerDb)
	mockLogger := hclog.NewNullLogger()
	config := GetDummyConfig()

	mockTxWritter := NewMockDbTransactionWriter()
	dummyTx := GetDummyShelleyTransaction_WithInterest()

	blockIndexer := NewBlockIndexer(config, nil, mockBlockSyncer, mockDb, mockLogger)

	blockIndexer.addAllTxOutputsToDb(mockTxWritter, []ledger.Transaction{dummyTx})
	mockTxWritter.Execute()

	// Assertions
	// TODO:
}

///////////////////////////////////////////////////

func GetDummyConfig() *BlockIndexerConfig {
	return &BlockIndexerConfig{
		NetworkMagic:           1,
		NodeAddress:            "preprod-node.play.dev.cardano.org:3001",
		StartingBlockPoint:     nil,
		ConfirmationBlockCount: 0,
		AddressesOfInterest: []string{
			"addr1qxaww0anyepl07pzdfm64pfk6xcm54kputjhnmqa9ku0d67jj9djsz0020h68nz3rxknzdh93nryqzhq6h9z0nnzf0rsfus4er",
			"addr1q900xuw7xruv836lks0hjyymwy9ft42x5mz0u387t9k3g50wv7ptw26llr0alv8n875rc8fw9ljyz5pxzxl8hgg8g3csdwghnl",
			// "addr1wygyy4mdrh5kxsmm6ja4phxez3vh2mngz2muqmw4gw3n9jqdu67a0",
		},
	}
}

func GetDummyShelleyTransaction_WithInterest() *ledger.ShelleyTransaction {
	address, _ := ledger.NewAddress("addr1qxaww0anyepl07pzdfm64pfk6xcm54kputjhnmqa9ku0d67jj9djsz0020h68nz3rxknzdh93nryqzhq6h9z0nnzf0rsfus4er")

	return &ledger.ShelleyTransaction{
		Body: ledger.ShelleyTransactionBody{
			TxOutputs: []ledger.ShelleyTransactionOutput{
				{
					OutputAddress: address,
					OutputAmount:  1000,
				},
			},
			TxInputs: []ledger.ShelleyTransactionInput{
				{
					// TODO: make tx id
					// prefix + bytes(data)
					TxId:        ledger.NewBlake2b256([]byte{}),
					OutputIndex: 1,
				},
			},
		},
	}
}

func GetDummyShelleyTransaction_WithoutInterest() *ledger.ShelleyTransaction {
	address, _ := ledger.NewAddress("addr1wygyy4mdrh5kxsmm6ja4phxez3vh2mngz2muqmw4gw3n9jqdu67a0")

	return &ledger.ShelleyTransaction{
		Body: ledger.ShelleyTransactionBody{
			TxOutputs: []ledger.ShelleyTransactionOutput{
				{
					OutputAddress: address,
					OutputAmount:  1000,
				},
			},
			TxInputs: []ledger.ShelleyTransactionInput{
				{
					// TODO: make tx id
					TxId:        ledger.NewBlake2b256([]byte{}),
					OutputIndex: 1,
				},
			},
		},
	}
}

func GetDummyBlockPoint(slot uint64, hash_ext string, block uint64) *BlockPoint {
	return &BlockPoint{
		BlockSlot:   slot,
		BlockHash:   []byte("34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d" + hash_ext),
		BlockNumber: block,
	}
}

func GetDummyBlockHeader(slot uint64, hash_ext string, block uint64) *BlockHeader {
	return &BlockHeader{
		BlockSlot:   slot,
		BlockHash:   []byte("34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d" + hash_ext),
		BlockNumber: block,
		EraID:       2,
		EraName:     "DummyEra",
	}
}

func GetDummyFullBlock(slot uint64, hash_ext string, block uint64) *FullBlock {
	return &FullBlock{
		BlockSlot:   slot,
		BlockHash:   []byte("34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d" + hash_ext),
		BlockNumber: block,
		EraID:       2,
		EraName:     "DummyEra",
		Txs: []*Tx{
			{
				Hash:     "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d20",
				Metadata: []byte("dummy_metadata"),
				Inputs: []*TxInput{
					{
						Hash:  "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d21",
						Index: 0,
					},
					{
						Hash:  "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d22",
						Index: 1,
					},
				},
				Outputs: []*TxOutput{
					{
						Address: "addr1qxaww0anyepl07pzdfm64pfk6xcm54kputjhnmqa9ku0d67jj9djsz0020h68nz3rxknzdh93nryqzhq6h9z0nnzf0rsfus4er",
						Amount:  1000,
					},
					{
						Address: "dummy_addr2_notinterested",
						Amount:  2000,
					},
				},
				Fee: 123,
			},
			{
				Hash:     "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d23",
				Metadata: []byte("dummy_metadata"),
				Inputs: []*TxInput{
					{
						Hash:  "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d24",
						Index: 0,
					},
					{
						Hash:  "34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d25",
						Index: 1,
					},
				},
				Outputs: []*TxOutput{
					{
						Address: "dummy_addr3_notinterested",
						Amount:  1000,
					},
					{
						Address: "dummy_addr4_notinterested",
						Amount:  2000,
					},
				},
				Fee: 123,
			},
		},
	}
}
