package core

type DbTransactionWriter interface {
	SetLatestBlockPoint(point *BlockPoint) DbTransactionWriter
	AddTxOutputs(txOutputs []*TxInputOutput) DbTransactionWriter
	AddConfirmedBlock(block *FullBlock) DbTransactionWriter
	RemoveTxOutputs(txInputs []*TxInput, softDelete bool) DbTransactionWriter
	Execute() error
}

type BlockIndexerDb interface {
	OpenTx() DbTransactionWriter
	GetTxOutput(txInput TxInput) (*TxOutput, error)
	GetLatestBlockPoint() (*BlockPoint, error)
}

type Database interface {
	BlockIndexerDb
	Init(filepath string) error

	MarkConfirmedBlockProcessed(block *FullBlock) error
	GetUnprocessedConfirmedBlocks() ([]*FullBlock, error)
}
