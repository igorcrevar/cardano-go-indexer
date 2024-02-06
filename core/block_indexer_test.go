package core

import (
	"testing"

	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

/*
Addresses
addr1vxrmu3m2cc5k6xltupj86a2uzcuq8r4nhznrhfq0pkwl4hgqj2v8w
addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd
addr1v8hrxaz0yqkfdsszfvjmdnqh0tv4xl2xgd7dfrxzj86cqzghu5c6p
addr1qxh7y2ezyt7hcraew7q0s8fg36usm049ktf4m9rly220snm0tf3rte5f4wequeg86kww58hp34qpwxdpl76tfuwmk77qjstmmj
*/

func TestBlockIndexer_processConfirmedBlockNoTxOfInterest(t *testing.T) {
	const (
		blockNumber = uint64(50)
		blockSlot   = uint64(100)
	)

	addressesOfInterest := []string{"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd"}
	blockHash := []byte{100, 200, 100}
	expectedLastBlockPoint := &BlockPoint{
		BlockSlot:   blockSlot,
		BlockHash:   blockHash,
		BlockNumber: blockNumber,
	}
	config := &BlockIndexerConfig{
		AddressCheck:        AddressCheckAll,
		AddressesOfInterest: addressesOfInterest,
	}
	dbMock := &DatabaseMock{
		Writter: &DbTransactionWriterMock{},
	}
	syncerMock := &BlockSyncerMock{}
	newConfirmedBlockHandler := func(fb *FullBlock) error {
		return nil
	}

	syncerMock.On("GetFullBlock", blockSlot, blockHash).Return(&LedgerBlockMock{
		TransactionsVal: []ledger.Transaction{
			&LedgerTransactionMock{
				InputsVal: []ledger.TransactionInput{
					NewLedgerTransactionInputMock(t, []byte{1, 2}, uint32(0)),
				},
				OutputsVal: []ledger.TransactionOutput{
					NewLedgerTransactionOutputMock(t, "addr1v8hrxaz0yqkfdsszfvjmdnqh0tv4xl2xgd7dfrxzj86cqzghu5c6p", uint64(100)),
				},
			},
			&LedgerTransactionMock{
				InputsVal: []ledger.TransactionInput{
					NewLedgerTransactionInputMock(t, []byte{1, 2}, uint32(1)),
					NewLedgerTransactionInputMock(t, []byte{1, 2, 3}, uint32(1)),
				},
				OutputsVal: []ledger.TransactionOutput{
					NewLedgerTransactionOutputMock(t, "addr1vxrmu3m2cc5k6xltupj86a2uzcuq8r4nhznrhfq0pkwl4hgqj2v8w", uint64(100)),
				},
			},
		},
	}, error(nil)).Once()
	dbMock.On("OpenTx").Once()
	dbMock.On("GetTxOutput", mock.Anything).Return((*TxOutput)(nil), error(nil)).Times(3)
	dbMock.Writter.On("Execute").Return(error(nil)).Once()
	dbMock.Writter.On("SetLatestBlockPoint", expectedLastBlockPoint).Once()
	dbMock.Writter.On("RemoveTxOutputs", ([]*TxInput)(nil)).Once()
	dbMock.Writter.On("AddTxOutputs", ([]*TxInputOutput)(nil)).Once()

	blockIndexer := NewBlockIndexer(config, newConfirmedBlockHandler, syncerMock, dbMock, hclog.NewNullLogger())
	assert.NotNil(t, blockIndexer)

	fb, latestBlockPoint, err := blockIndexer.processConfirmedBlock(&BlockHeader{
		BlockSlot:   blockSlot,
		BlockHash:   blockHash,
		BlockNumber: blockNumber,
	})

	require.Nil(t, err)
	assert.Nil(t, fb)
	assert.Equal(t, expectedLastBlockPoint, latestBlockPoint)
	syncerMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	dbMock.Writter.AssertExpectations(t)
}

func TestBlockIndexer_processConfirmedBlockTxOfInterestInOutputs(t *testing.T) {
	const (
		blockNumber = uint64(50)
		blockSlot   = uint64(100)
	)

	hashTx := []string{"00333", "7873282"}
	addressesOfInterest := []string{"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd", "addr1qxh7y2ezyt7hcraew7q0s8fg36usm049ktf4m9rly220snm0tf3rte5f4wequeg86kww58hp34qpwxdpl76tfuwmk77qjstmmj"}
	blockHash := []byte{100, 200, 100}
	txInputs := [3]ledger.TransactionInput{
		NewLedgerTransactionInputMock(t, []byte{1}, uint32(0)),
		NewLedgerTransactionInputMock(t, []byte{1, 2}, uint32(1)),
		NewLedgerTransactionInputMock(t, []byte{1, 2, 3}, uint32(2)),
	}
	txOutputs := [4]ledger.TransactionOutput{
		NewLedgerTransactionOutputMock(t, addressesOfInterest[0], uint64(100)),
		NewLedgerTransactionOutputMock(t, addressesOfInterest[1], uint64(200)),
		NewLedgerTransactionOutputMock(t, "addr1vxrmu3m2cc5k6xltupj86a2uzcuq8r4nhznrhfq0pkwl4hgqj2v8w", uint64(100)), // not address of interest
		NewLedgerTransactionOutputMock(t, "addr1vxrmu3m2cc5k6xltupj86a2uzcuq8r4nhznrhfq0pkwl4hgqj2v8w", uint64(100)), // not address of interest
	}
	expectedLastBlockPoint := &BlockPoint{
		BlockSlot:   blockSlot,
		BlockHash:   blockHash,
		BlockNumber: blockNumber,
	}
	config := &BlockIndexerConfig{
		AddressCheck:        AddressCheckAll,
		AddressesOfInterest: addressesOfInterest,
	}
	dbMock := &DatabaseMock{
		Writter: &DbTransactionWriterMock{},
	}
	syncerMock := &BlockSyncerMock{}
	newConfirmedBlockHandler := func(fb *FullBlock) error {
		return nil
	}

	syncerMock.On("GetFullBlock", blockSlot, blockHash).Return(&LedgerBlockMock{
		TransactionsVal: []ledger.Transaction{
			&LedgerTransactionMock{
				HashVal: hashTx[0],
				InputsVal: []ledger.TransactionInput{
					txInputs[0],
				},
				OutputsVal: []ledger.TransactionOutput{
					txOutputs[0],
				},
			},
			&LedgerTransactionMock{
				InputsVal: []ledger.TransactionInput{
					txInputs[1],
				},
				OutputsVal: []ledger.TransactionOutput{
					txOutputs[3],
				},
			},
			&LedgerTransactionMock{
				HashVal: hashTx[1],
				InputsVal: []ledger.TransactionInput{
					txInputs[2],
				},
				OutputsVal: []ledger.TransactionOutput{
					txOutputs[2],
					txOutputs[1],
				},
			},
		},
	}, error(nil)).Once()
	dbMock.On("OpenTx").Once()
	// one call will be for address of interest inside inputs
	dbMock.On("GetTxOutput", TxInput{
		Hash:  txInputs[1].Id().String(),
		Index: txInputs[1].Index(),
	}).Return((*TxOutput)(nil), error(nil)).Once()
	dbMock.Writter.On("Execute").Return(error(nil)).Once()
	dbMock.Writter.On("SetLatestBlockPoint", expectedLastBlockPoint).Once()
	dbMock.Writter.On("RemoveTxOutputs", []*TxInput{
		{
			Hash:  txInputs[0].Id().String(),
			Index: txInputs[0].Index(),
		},
		{
			Hash:  txInputs[2].Id().String(),
			Index: txInputs[2].Index(),
		},
	}).Once()
	dbMock.Writter.On("AddTxOutputs", []*TxInputOutput{
		{
			Input: &TxInput{
				Hash:  hashTx[0],
				Index: 0,
			},
			Output: &TxOutput{
				Address: txOutputs[0].Address().String(),
				Amount:  txOutputs[0].Amount(),
			},
		},
		{
			Input: &TxInput{
				Hash:  hashTx[1],
				Index: 1,
			},
			Output: &TxOutput{
				Address: txOutputs[1].Address().String(),
				Amount:  txOutputs[1].Amount(),
			},
		},
	}).Once()
	dbMock.Writter.On("AddConfirmedBlock", mock.Anything).Run(func(args mock.Arguments) {
		block := args.Get(0).(*FullBlock)
		require.NotNil(t, block)
		require.Equal(t, block.BlockHash, blockHash)
		require.Len(t, block.Txs, 2)
	}).Once()

	blockIndexer := NewBlockIndexer(config, newConfirmedBlockHandler, syncerMock, dbMock, hclog.NewNullLogger())
	assert.NotNil(t, blockIndexer)

	fb, latestBlockPoint, err := blockIndexer.processConfirmedBlock(&BlockHeader{
		BlockSlot:   blockSlot,
		BlockHash:   blockHash,
		BlockNumber: blockNumber,
	})

	require.Nil(t, err)
	require.NotNil(t, fb)
	require.Len(t, fb.Txs, 2)
	assert.Equal(t, fb.Txs[0].Hash, hashTx[0])
	assert.Equal(t, fb.Txs[1].Hash, hashTx[1])
	assert.Equal(t, expectedLastBlockPoint, latestBlockPoint)
	syncerMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	dbMock.Writter.AssertExpectations(t)
}

func TestBlockIndexer_processConfirmedBlockTxOfInterestInInputs(t *testing.T) {
	const (
		blockNumber = uint64(50)
		blockSlot   = uint64(100)
	)

	hashTx := [2]string{"eee", "111"}
	addressesOfInterest := []string{"addr1v9kganeshgdqyhwnyn9stxxgl7r4y2ejfyqjn88n7ncapvs4sugsd", "addr1qxh7y2ezyt7hcraew7q0s8fg36usm049ktf4m9rly220snm0tf3rte5f4wequeg86kww58hp34qpwxdpl76tfuwmk77qjstmmj"}
	dbInputOutputs := [2]*TxInputOutput{
		{
			Input: &TxInput{
				Hash:  string("xyzy"),
				Index: uint32(20),
			},
			Output: &TxOutput{
				Address: addressesOfInterest[0],
				Amount:  2000,
			},
		},
		{
			Input: &TxInput{
				Hash:  string("abcdef"),
				Index: uint32(120),
			},
			Output: &TxOutput{
				Address: addressesOfInterest[1],
				Amount:  2,
			},
		},
	}
	txInputs := [4]*LedgerTransactionInputMock{
		NewLedgerTransactionInputMock(t, []byte("not_exist_1"), uint32(0)),
		NewLedgerTransactionInputMock(t, []byte(dbInputOutputs[0].Input.Hash), dbInputOutputs[0].Input.Index),
		NewLedgerTransactionInputMock(t, []byte("not_exist_2"), uint32(0)),
		NewLedgerTransactionInputMock(t, []byte(dbInputOutputs[1].Input.Hash), dbInputOutputs[1].Input.Index),
	}
	blockHash := []byte{100, 200, 100}
	expectedLastBlockPoint := &BlockPoint{
		BlockSlot:   blockSlot,
		BlockHash:   blockHash,
		BlockNumber: blockNumber,
	}
	config := &BlockIndexerConfig{
		AddressCheck:        AddressCheckAll,
		AddressesOfInterest: addressesOfInterest,
	}
	dbMock := &DatabaseMock{
		Writter: &DbTransactionWriterMock{},
	}
	syncerMock := &BlockSyncerMock{}
	newConfirmedBlockHandler := func(fb *FullBlock) error {
		return nil
	}

	syncerMock.On("GetFullBlock", blockSlot, blockHash).Return(&LedgerBlockMock{
		TransactionsVal: []ledger.Transaction{
			&LedgerTransactionMock{
				HashVal: hashTx[0],
				InputsVal: []ledger.TransactionInput{
					txInputs[0], txInputs[1],
				},
				OutputsVal: []ledger.TransactionOutput{
					NewLedgerTransactionOutputMock(t, "addr1vxrmu3m2cc5k6xltupj86a2uzcuq8r4nhznrhfq0pkwl4hgqj2v8w", uint64(200)),
				},
			},
			&LedgerTransactionMock{
				InputsVal: []ledger.TransactionInput{
					txInputs[2],
				},
				OutputsVal: []ledger.TransactionOutput{
					NewLedgerTransactionOutputMock(t, "addr1vxrmu3m2cc5k6xltupj86a2uzcuq8r4nhznrhfq0pkwl4hgqj2v8w", uint64(100)),
				},
			},
			&LedgerTransactionMock{
				HashVal: hashTx[1],
				InputsVal: []ledger.TransactionInput{
					txInputs[3],
				},
				OutputsVal: []ledger.TransactionOutput{
					NewLedgerTransactionOutputMock(t, "addr1vxrmu3m2cc5k6xltupj86a2uzcuq8r4nhznrhfq0pkwl4hgqj2v8w", uint64(200)),
				},
			},
		},
	}, error(nil)).Once()
	dbMock.On("OpenTx").Once()
	dbMock.On("GetTxOutput", TxInput{
		Hash:  txInputs[0].Id().String(),
		Index: txInputs[0].Index(),
	}).Return((*TxOutput)(nil), error(nil)).Once()
	dbMock.On("GetTxOutput", TxInput{
		Hash:  txInputs[2].Id().String(),
		Index: txInputs[2].Index(),
	}).Return((*TxOutput)(nil), error(nil)).Once()
	dbMock.On("GetTxOutput", TxInput{
		Hash:  txInputs[1].Id().String(),
		Index: txInputs[1].Index(),
	}).Return(&TxOutput{
		Address: addressesOfInterest[0],
	}, error(nil)).Once()
	dbMock.On("GetTxOutput", TxInput{
		Hash:  txInputs[3].Id().String(),
		Index: txInputs[3].Index(),
	}).Return(&TxOutput{
		Address: addressesOfInterest[1],
	}, error(nil)).Once()
	dbMock.Writter.On("Execute").Return(error(nil)).Once()
	dbMock.Writter.On("SetLatestBlockPoint", expectedLastBlockPoint).Once()
	dbMock.Writter.On("RemoveTxOutputs", []*TxInput{
		{
			Hash:  txInputs[0].Id().String(),
			Index: txInputs[0].Index(),
		},
		{
			Hash:  txInputs[1].Id().String(),
			Index: txInputs[1].Index(),
		},
		{
			Hash:  txInputs[3].Id().String(),
			Index: txInputs[3].Index(),
		},
	}).Once()
	dbMock.Writter.On("AddTxOutputs", ([]*TxInputOutput)(nil)).Once()
	dbMock.Writter.On("AddConfirmedBlock", mock.Anything).Run(func(args mock.Arguments) {
		block := args.Get(0).(*FullBlock)
		require.NotNil(t, block)
		require.Equal(t, block.BlockHash, blockHash)
		require.Len(t, block.Txs, 2)
	}).Once()

	blockIndexer := NewBlockIndexer(config, newConfirmedBlockHandler, syncerMock, dbMock, hclog.NewNullLogger())
	assert.NotNil(t, blockIndexer)

	fb, latestBlockPoint, err := blockIndexer.processConfirmedBlock(&BlockHeader{
		BlockSlot:   blockSlot,
		BlockHash:   blockHash,
		BlockNumber: blockNumber,
	})

	require.Nil(t, err)
	require.NotNil(t, fb)
	require.Len(t, fb.Txs, 2)
	assert.Equal(t, fb.Txs[0].Hash, hashTx[0])
	assert.Equal(t, fb.Txs[1].Hash, hashTx[1])
	assert.Equal(t, expectedLastBlockPoint, latestBlockPoint)
	syncerMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
	dbMock.Writter.AssertExpectations(t)
}
