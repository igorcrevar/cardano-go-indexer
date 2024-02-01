package core

import (
	"testing"

	"github.com/blinklabs-io/gouroboros/ledger"
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
	return new(MockDbTransactionWriter)
}

func (m *MockBlockIndexerDb) GetTxOutput(input TxInput) (*TxOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*TxOutput), args.Error(1)
}

// MockDbTransactionWriter
type MockDbTransactionWriter struct {
	mock.Mock
	Blocks  []*FullBlock
	TxInOut map[TxInput]TxOutput
}

func NewMockDbTransactionWriter() *MockDbTransactionWriter {
	mockDbTxWriter := new(MockDbTransactionWriter)
	mockDbTxWriter.TxInOut = make(map[TxInput]TxOutput)
	return mockDbTxWriter
}

func (m *MockDbTransactionWriter) AddConfirmedBlock(block *FullBlock) DbTransactionWriter {
	m.Blocks = append(m.Blocks, block)
	return m
}

func (m *MockDbTransactionWriter) SetLatestBlockPoint(point *BlockPoint) DbTransactionWriter {
	m.Called(point)
	return m
}

func (m *MockDbTransactionWriter) AddTxOutput(input TxInput, output *TxOutput) DbTransactionWriter {
	m.TxInOut[input] = *output
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

	//mockDb.On("GetLatestBlockPoint").Return(GetDummyBlockPoint(), nil)

	// Assertions
	assert.NotNil(t, blockIndexer)
	assert.Equal(t, config, blockIndexer.config)
	assert.NotNil(t, blockIndexer.addressesOfInterest)
	assert.Equal(t, len(blockIndexer.addressesOfInterest), 2)
	assert.NotNil(t, blockIndexer.logger)
}

func TestRollBackwardFunc(t *testing.T) {
	assert.Equal(t, 1, 1)
	// TODO: Implement RollBackwardFunc test cases
}

// TODO get blockinfo and tip objects from cardano chain
func TestRollForwardFunc(t *testing.T) {
	mockBlockSyncer := new(MockBlockSyncer)
	mockDb := new(MockBlockIndexerDb)
	mockLogger := hclog.NewNullLogger()
	config := GetDummyConfig()

	blockIndexer := NewBlockIndexer(config, nil, mockBlockSyncer, mockDb, mockLogger)

	// blockIndexer.RollForwardFunc(7, nil, nil)

	assert.NotNil(t, blockIndexer.unconfirmedBlocks)
	assert.Equal(t, len(blockIndexer.unconfirmedBlocks), 1)
	assert.EqualValues(t, blockIndexer.unconfirmedBlocks[0], &BlockHeader{})

	// BlockNumber
	// BlockHash

	// Hash
	// BlockNumber
	// SlotNumber
	// Era

	// my BlockHeader
	// bi.unconfirmed blocks da ima my BlockHeader
}

func TestProcessNewConfirmedBlock(t *testing.T) {
	// TODO: Implement ProcessNewConfirmedBlock test cases
}

func TestGetTxsOfInterest(t *testing.T) {
	// TODO: Implement GetTxsOfInterest test cases
}

func TestIsTxOutputOfInterest(t *testing.T) {
	// TODO: Implement IsTxOutputOfInterest test cases
}

func TestIsTxInputOfInterest(t *testing.T) {
	// TODO: Implement IsTxInputOfInterest test cases
}

func TestAddTxOutputs(t *testing.T) {
	mockBlockSyncer := new(MockBlockSyncer)
	mockDb := new(MockBlockIndexerDb)
	mockLogger := hclog.NewNullLogger()
	config := GetDummyConfig()

	mockTxWritter := NewMockDbTransactionWriter()
	dummyBlock := GetDummyFullBlock()

	blockIndexer := NewBlockIndexer(config, nil, mockBlockSyncer, mockDb, mockLogger)

	blockIndexer.addTxOutputs(mockTxWritter, dummyBlock)
	mockTxWritter.Execute()

	// Assertions
	// TODO:
}

///////////////////////////////////////////////////

func GetDummyConfig() *BlockIndexerConfig {
	return &BlockIndexerConfig{
		NetworkMagic:           9000,
		NodeAddress:            "localhost:3000",
		StartingBlockPoint:     nil,
		ConfirmationBlockCount: 10,
		AddressesOfInterest:    []string{"dummy_addr1", "dummy_addr2"},
	}
}

func GetDummyBlockPoint() *BlockPoint {
	return &BlockPoint{
		BlockSlot:   1,
		BlockHash:   []byte("34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d19"),
		BlockNumber: 1,
	}
}

func GetDummyBlockHeader() *BlockHeader {
	return &BlockHeader{
		BlockSlot:   1,
		BlockHash:   []byte("34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d19"),
		BlockNumber: 1,
		EraID:       2,
		EraName:     "DummyEra",
	}
}

func GetDummyFullBlock() *FullBlock {
	return &FullBlock{
		BlockSlot:   1,
		BlockHash:   []byte("34c36a9eb7228ca529e91babcf2215be29ce2a65b609540b483abc4520848d19"),
		BlockNumber: 1,
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
						Address: "dummy_addr1",
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
