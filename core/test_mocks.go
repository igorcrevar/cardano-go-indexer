package core

import (
	"testing"

	"github.com/blinklabs-io/gouroboros/cbor"
	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/ledger/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	utxorpc "github.com/utxorpc/go-codegen/utxorpc/v1alpha/cardano"
)

type BlockSyncerMock struct {
	mock.Mock
	CloseFn func() error
	SyncFn  func() error
}

// Close implements BlockSyncer.
func (m *BlockSyncerMock) Close() error {
	args := m.Called()

	if m.CloseFn != nil {
		return m.CloseFn()
	}

	return args.Error(0)
}

// Sync implements BlockSyncer.
func (m *BlockSyncerMock) Sync() error {
	args := m.Called()

	if m.SyncFn != nil {
		return m.SyncFn()
	}

	return args.Error(0)
}

// ErrorCh implements BlockSyncer.
func (m *BlockSyncerMock) ErrorCh() <-chan error {
	return make(<-chan error)
}

var _ BlockSyncer = (*BlockSyncerMock)(nil)

type DatabaseMock struct {
	mock.Mock
	Writter               *DBTransactionWriterMock
	GetLatestBlockPointFn func() (*BlockPoint, error)
	GetTxOutputFn         func(txInput TxInput) (TxOutput, error)
	InitFn                func(filepath string) error
}

// GetLatestBlockPoint implements Database.
func (m *DatabaseMock) GetLatestBlockPoint() (*BlockPoint, error) {
	args := m.Called()

	if m.GetLatestBlockPointFn != nil {
		return m.GetLatestBlockPointFn()
	}

	//nolint:forcetypeassert
	return args.Get(0).(*BlockPoint), args.Error(1)
}

// GetTxOutput implements Database.
func (m *DatabaseMock) GetTxOutput(txInput TxInput) (TxOutput, error) {
	args := m.Called(txInput)

	if m.GetTxOutputFn != nil {
		return m.GetTxOutputFn(txInput)
	}

	//nolint:forcetypeassert
	return args.Get(0).(TxOutput), args.Error(1)
}

// GetUnprocessedConfirmedTxs implements Database.
func (m *DatabaseMock) GetUnprocessedConfirmedTxs(maxCnt int) ([]*Tx, error) {
	args := m.Called(maxCnt)

	//nolint:forcetypeassert
	return args.Get(0).([]*Tx), args.Error(1)
}

// Init implements Database.
func (m *DatabaseMock) Init(filepath string) error {
	args := m.Called(filepath)

	if m.InitFn != nil {
		return m.InitFn(filepath)
	}

	return args.Error(0)
}

// MarkConfirmedTxsProcessed implements Database.
func (m *DatabaseMock) MarkConfirmedTxsProcessed(txs []*Tx) error {
	return m.Called(txs).Error(0)
}

// OpenTx implements Database.
func (m *DatabaseMock) OpenTx() DBTransactionWriter {
	args := m.Called()

	if m.Writter != nil {
		return m.Writter
	}

	//nolint:forcetypeassert
	return args.Get(0).(DBTransactionWriter)
}

func (m *DatabaseMock) Close() error {
	return m.Called().Error(0)
}

func (m *DatabaseMock) GetLatestConfirmedBlocks(maxCnt int) ([]*CardanoBlock, error) {
	args := m.Called(maxCnt)

	//nolint:forcetypeassert
	return args.Get(0).([]*CardanoBlock), args.Error(1)
}

func (m *DatabaseMock) GetConfirmedBlocksFrom(slotNumber uint64, maxCnt int) ([]*CardanoBlock, error) {
	args := m.Called(slotNumber, maxCnt)

	//nolint:forcetypeassert
	return args.Get(0).([]*CardanoBlock), args.Error(1)
}

func (m *DatabaseMock) GetAllTxOutputs(address string, onlyNotUser bool) ([]*TxInputOutput, error) {
	args := m.Called(address, onlyNotUser)

	//nolint:forcetypeassert
	return args.Get(0).([]*TxInputOutput), args.Error(1)
}

var _ Database = (*DatabaseMock)(nil)

type DBTransactionWriterMock struct {
	mock.Mock
	AddConfirmedTxsFn     func(txs []*Tx) DBTransactionWriter
	AddTxOutputsFn        func(txOutputs []*TxInputOutput) DBTransactionWriter
	RemoveTxOutputsFn     func(txInputs []*TxInput) DBTransactionWriter
	SetLatestBlockPointFn func(point *BlockPoint) DBTransactionWriter
	ExecuteFn             func() error
}

// AddConfirmedTxs implements DbTransactionWriter.
func (m *DBTransactionWriterMock) AddConfirmedTxs(txs []*Tx) DBTransactionWriter {
	m.Called(txs)

	if m.AddConfirmedTxsFn != nil {
		return m.AddConfirmedTxsFn(txs)
	}

	return m
}

func (m *DBTransactionWriterMock) AddConfirmedBlock(block *CardanoBlock) DBTransactionWriter {
	m.Called(block)

	return m
}

// AddTxOutputs implements DbTransactionWriter.
func (m *DBTransactionWriterMock) AddTxOutputs(txOutputs []*TxInputOutput) DBTransactionWriter {
	m.Called(txOutputs)

	if m.AddTxOutputsFn != nil {
		return m.AddTxOutputsFn(txOutputs)
	}

	return m
}

// Execute implements DbTransactionWriter.
func (m *DBTransactionWriterMock) Execute() error {
	if m.ExecuteFn != nil {
		return m.ExecuteFn()
	}

	return m.Called().Error(0)
}

// RemoveTxOutputs implements DbTransactionWriter.
func (m *DBTransactionWriterMock) RemoveTxOutputs(txInputs []*TxInput, softDelete bool) DBTransactionWriter {
	m.Called(txInputs, softDelete)

	if m.RemoveTxOutputsFn != nil {
		return m.RemoveTxOutputsFn(txInputs)
	}

	return m
}

// SetLatestBlockPoint implements DbTransactionWriter.
func (m *DBTransactionWriterMock) SetLatestBlockPoint(point *BlockPoint) DBTransactionWriter {
	m.Called(point)

	if m.SetLatestBlockPointFn != nil {
		return m.SetLatestBlockPointFn(point)
	}

	return m
}

func (m *DBTransactionWriterMock) DeleteAllTxOutputsPhysically() DBTransactionWriter {
	m.Called()

	return m
}

var _ DBTransactionWriter = (*DBTransactionWriterMock)(nil)

type LedgerBlockHeaderMock struct {
	BlockNumberVal uint64
	SlotNumberVal  uint64
	EraVal         ledger.Era
	HashVal        string
}

// BlockBodySize implements ledger.BlockHeader.
func (m *LedgerBlockHeaderMock) BlockBodySize() uint64 {
	panic("unimplemented") //nolint
}

// BlockNumber implements ledger.BlockHeader.
func (m *LedgerBlockHeaderMock) BlockNumber() uint64 {
	return m.BlockNumberVal
}

// Cbor implements ledger.BlockHeader.
func (m *LedgerBlockHeaderMock) Cbor() []byte {
	panic("unimplemented") //nolint
}

// Era implements ledger.BlockHeader.
func (m *LedgerBlockHeaderMock) Era() ledger.Era {
	return m.EraVal
}

// Hash implements ledger.BlockHeader.
func (m *LedgerBlockHeaderMock) Hash() string {
	return m.HashVal
}

// IssuerVkey implements ledger.BlockHeader.
func (m *LedgerBlockHeaderMock) IssuerVkey() ledger.IssuerVkey {
	panic("unimplemented") //nolint
}

// SlotNumber implements ledger.BlockHeader.
func (m *LedgerBlockHeaderMock) SlotNumber() uint64 {
	return m.SlotNumberVal
}

var _ ledger.BlockHeader = (*LedgerBlockHeaderMock)(nil)

type LedgerBlockMock struct {
	TransactionsVal []ledger.Transaction
}

// Type implements ledger.Block.
func (m *LedgerBlockMock) Type() int {
	panic("unimplemented") //nolint
}

// BlockBodySize implements ledger.Block.
func (m *LedgerBlockMock) BlockBodySize() uint64 {
	panic("unimplemented") //nolint
}

// BlockNumber implements ledger.Block.
func (*LedgerBlockMock) BlockNumber() uint64 {
	panic("unimplemented") //nolint
}

// Cbor implements ledger.Block.
func (m *LedgerBlockMock) Cbor() []byte {
	panic("unimplemented") //nolint
}

// Era implements ledger.Block.
func (m *LedgerBlockMock) Era() ledger.Era {
	panic("unimplemented") //nolint
}

// Hash implements ledger.Block.
func (m *LedgerBlockMock) Hash() string {
	panic("unimplemented") //nolint
}

// IssuerVkey implements ledger.Block.
func (m *LedgerBlockMock) IssuerVkey() ledger.IssuerVkey {
	panic("unimplemented") //nolint
}

// SlotNumber implements ledger.Block.
func (m *LedgerBlockMock) SlotNumber() uint64 {
	panic("unimplemented") //nolint
}

// Transactions implements ledger.Block.
func (m *LedgerBlockMock) Transactions() []ledger.Transaction {
	return m.TransactionsVal
}

func (m *LedgerBlockMock) Utxorpc() *utxorpc.Block {
	return nil
}

var _ ledger.Block = (*LedgerBlockMock)(nil)

type LedgerTransactionMock struct {
	FeeVal             uint64
	HashVal            string
	InputsVal          []ledger.TransactionInput
	OutputsVal         []ledger.TransactionOutput
	MetadataVal        *cbor.LazyValue
	TTLVal             uint64
	IsInvalid          bool
	ReferenceInputsVal []ledger.TransactionInput
}

// AssetMint implements common.Transaction.
func (m *LedgerTransactionMock) AssetMint() *common.MultiAsset[int64] {
	panic("unimplemented") //nolint
}

// AuxDataHash implements common.Transaction.
func (m *LedgerTransactionMock) AuxDataHash() *common.Blake2b256 {
	panic("unimplemented") //nolint
}

// Certificates implements common.Transaction.
func (m *LedgerTransactionMock) Certificates() []common.Certificate {
	panic("unimplemented") //nolint
}

// Collateral implements common.Transaction.
func (m *LedgerTransactionMock) Collateral() []common.TransactionInput {
	panic("unimplemented") //nolint
}

// CollateralReturn implements common.Transaction.
func (m *LedgerTransactionMock) CollateralReturn() common.TransactionOutput {
	panic("unimplemented") //nolint
}

// Consumed implements common.Transaction.
func (m *LedgerTransactionMock) Consumed() []common.TransactionInput {
	panic("unimplemented") //nolint
}

// CurrentTreasuryValue implements common.Transaction.
func (m *LedgerTransactionMock) CurrentTreasuryValue() int64 {
	panic("unimplemented") //nolint
}

// Donation implements common.Transaction.
func (m *LedgerTransactionMock) Donation() uint64 {
	panic("unimplemented") //nolint
}

// Produced implements common.Transaction.
func (m *LedgerTransactionMock) Produced() []common.Utxo {
	panic("unimplemented") //nolint
}

// ProposalProcedures implements common.Transaction.
func (m *LedgerTransactionMock) ProposalProcedures() []common.ProposalProcedure {
	panic("unimplemented") //nolint
}

// ProtocolParameterUpdates implements common.Transaction.
func (m *LedgerTransactionMock) ProtocolParameterUpdates() (
	uint64, map[common.Blake2b224]common.ProtocolParameterUpdate,
) {
	panic("unimplemented") //nolint
}

// RequiredSigners implements common.Transaction.
func (m *LedgerTransactionMock) RequiredSigners() []common.Blake2b224 {
	panic("unimplemented") //nolint
}

// ScriptDataHash implements common.Transaction.
func (m *LedgerTransactionMock) ScriptDataHash() *common.Blake2b256 {
	panic("unimplemented") //nolint
}

// TotalCollateral implements common.Transaction.
func (m *LedgerTransactionMock) TotalCollateral() uint64 {
	panic("unimplemented") //nolint
}

// Type implements common.Transaction.
func (m *LedgerTransactionMock) Type() int {
	panic("unimplemented") //nolint
}

// ValidityIntervalStart implements common.Transaction.
func (m *LedgerTransactionMock) ValidityIntervalStart() uint64 {
	panic("unimplemented") //nolint
}

// VotingProcedures implements common.Transaction.
func (m *LedgerTransactionMock) VotingProcedures() common.VotingProcedures {
	panic("unimplemented") //nolint
}

// Withdrawals implements common.Transaction.
func (m *LedgerTransactionMock) Withdrawals() map[*common.Address]uint64 {
	panic("unimplemented") //nolint
}

// Cbor implements ledger.Transaction.
func (m *LedgerTransactionMock) Cbor() []byte {
	panic("unimplemented") //nolint
}

// Fee implements ledger.Transaction.
func (m *LedgerTransactionMock) Fee() uint64 {
	return m.FeeVal
}

// Hash implements ledger.Transaction.
func (m *LedgerTransactionMock) Hash() string {
	return m.HashVal
}

// Inputs implements ledger.Transaction.
func (m *LedgerTransactionMock) Inputs() []ledger.TransactionInput {
	return m.InputsVal
}

// Metadata implements ledger.Transaction.
func (m *LedgerTransactionMock) Metadata() *cbor.LazyValue {
	return m.MetadataVal
}

// Outputs implements ledger.Transaction.
func (m *LedgerTransactionMock) Outputs() []ledger.TransactionOutput {
	return m.OutputsVal
}

// TTL implements ledger.Transaction.
func (m *LedgerTransactionMock) TTL() uint64 {
	return m.TTLVal
}

func (m *LedgerTransactionMock) Utxorpc() *utxorpc.Tx {
	return nil
}

func (m *LedgerTransactionMock) IsValid() bool {
	return !m.IsInvalid
}

func (m *LedgerTransactionMock) ReferenceInputs() []ledger.TransactionInput {
	return m.ReferenceInputsVal
}

var _ ledger.Transaction = (*LedgerTransactionMock)(nil)

type LedgerTransactionInputMock struct {
	HashVal  ledger.Blake2b256
	IndexVal uint32
}

func NewLedgerTransactionInputMock(t *testing.T, hash []byte, index uint32) *LedgerTransactionInputMock {
	t.Helper()

	return &LedgerTransactionInputMock{
		HashVal:  ledger.NewBlake2b256(hash),
		IndexVal: index,
	}
}

// Id implements ledger.TransactionInput.
func (m *LedgerTransactionInputMock) Id() ledger.Blake2b256 { //nolint
	return m.HashVal
}

// Index implements ledger.TransactionInput.
func (m *LedgerTransactionInputMock) Index() uint32 {
	return m.IndexVal
}

func (m *LedgerTransactionInputMock) Utxorpc() *utxorpc.TxInput {
	return nil
}

var _ ledger.TransactionInput = (*LedgerTransactionInputMock)(nil)

type LedgerTransactionOutputMock struct {
	AddressVal   ledger.Address
	AmountVal    uint64
	DatumVal     *cbor.LazyValue
	DatumHashVal *common.Blake2b256
	AssetsVal    *common.MultiAsset[uint64]
}

// Assets implements common.TransactionOutput.
func (m *LedgerTransactionOutputMock) Assets() *common.MultiAsset[uint64] {
	return m.AssetsVal
}

// Cbor implements common.TransactionOutput.
func (m *LedgerTransactionOutputMock) Cbor() []byte {
	panic("unimplemented") //nolint
}

// Datum implements common.TransactionOutput.
func (m *LedgerTransactionOutputMock) Datum() *cbor.LazyValue {
	return m.DatumVal
}

// DatumHash implements common.TransactionOutput.
func (m *LedgerTransactionOutputMock) DatumHash() *common.Blake2b256 {
	return m.DatumHashVal
}

// Address implements ledger.TransactionOutput.
func (m *LedgerTransactionOutputMock) Address() ledger.Address {
	return m.AddressVal
}

// Amount implements ledger.TransactionOutput.
func (m *LedgerTransactionOutputMock) Amount() uint64 {
	return m.AmountVal
}

func (m *LedgerTransactionOutputMock) Utxorpc() *utxorpc.TxOutput {
	return nil
}

func NewLedgerTransactionOutputMock(t *testing.T, addr string, amount uint64) *LedgerTransactionOutputMock {
	t.Helper()

	a, err := ledger.NewAddress(addr)
	require.NoError(t, err)

	return &LedgerTransactionOutputMock{
		AddressVal: a,
		AmountVal:  amount,
	}
}

var _ ledger.TransactionOutput = (*LedgerTransactionOutputMock)(nil)
