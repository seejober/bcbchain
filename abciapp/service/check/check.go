package check

import (
	"fmt"
	types3 "github.com/bcbchain/bcbchain/abciapp/service/types"
	"github.com/bcbchain/bclib/tendermint/abci/types"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
	"github.com/bcbchain/bclib/tendermint/tmlibs/log"
	types2 "github.com/bcbchain/bclib/types"
	"sync"
)

//AppCheck object of check tx
//nolint
type AppCheck struct {
	logger  log.Logger
	chainID string
}

//SetLogger set logger
func (app *AppCheck) SetLogger(logger log.Logger) {
	app.logger = logger
}

//SetChainID set chainID
func (app *AppCheck) SetChainID(chainID string) {
	app.chainID = chainID
}

//CheckTx check tx
func (app *AppCheck) CheckTx(tx []byte) types.ResponseCheckTx {
	app.logger.Info("Recv ABCI interface: CheckTx", "tx", string(tx))

	return app.CheckBCTx(tx)
}

//CheckTxV2Concurrency check tx v2
func (app *AppCheck) CheckTxV2Concurrency(tx []byte, wg sync.WaitGroup, resultChan *types3.ResultPool, index int) {
	defer wg.Done()
	app.logger.Info("Recv ABCI interface: CheckTxV2Concurrency", "tx", string(tx))

	result := app.CheckBCTxV2Concurrency(tx, wg)
	result.TxID = index
	resultChan.ResultChan <- *result
}

//CheckTxV2Concurrency check tx v3
func (app *AppCheck) CheckTxV3Concurrency(tx []byte, wg sync.WaitGroup, resultChan *types3.ResultPool, index int) {
	defer wg.Done()
	app.logger.Info("Recv ABCI interface: CheckTxV2Concurrency", "tx", string(tx))

	result := app.CheckBCTxV3Concurrency(tx, wg)
	result.TxID = index
	resultChan.ResultChan <- *result
}

// ------------- add for support v1 transaction begin ----------------

//RunCheckTx - invoked by v1 checkTx, if it's standard transfer method.
func (app *AppCheck) RunCheckTx(tx []byte, transaction types2.Transaction, pubKey crypto.PubKeyEd25519) types.ResponseCheckTx {
	app.logger.Debug("Recv ABCI interface: CheckTx", "transaction", transaction)

	return app.runCheckBCTx(tx, transaction, pubKey)
}

func (app *AppCheck) RunCheckTxV2Concurrency(result types.Result, responsePool *types3.ResponsePool, wg sync.WaitGroup) {
	defer wg.Done()
	app.logger.Debug("Recv ABCI interface: CheckTxConcurrency", "transaction", result.TxV2Result.Transaction)

	if result.Errorlog != nil {
		responseChanOrder := types3.ResponseChanOrder{
			Response: types.ResponseCheckTx{
				Code: types2.ErrCheckTx,
				Log:  fmt.Sprint(result.Errorlog),
			},
			Index: result.TxID,
		}
		//responsePool.Response <- responseCheckTx
		responsePool.ResponseOrder <- responseChanOrder
	}
	responseCheckTx := app.runCheckBCTxV2Concurrency(result)
	responseChanOrder := types3.ResponseChanOrder{
		Response: responseCheckTx,
		Index:    result.TxID,
	}
	//responsePool.Response <- responseCheckTx
	responsePool.ResponseOrder <- responseChanOrder

}

func (app *AppCheck) RunCheckTxV3Concurrency(result types.Result, responsePool *types3.ResponsePool, wg sync.WaitGroup) {
	defer wg.Done()
	app.logger.Debug("Recv ABCI interface: CheckTx", "transaction", result.TxV3Result.Transaction)

	if result.Errorlog != nil {
		responseChanOrder := types3.ResponseChanOrder{
			Response: types.ResponseCheckTx{
				Code: types2.ErrCheckTx,
				Log:  fmt.Sprint(result.Errorlog),
			},
			Index: result.TxID,
		}
		//responsePool.Response <- responseCheckTx
		responsePool.ResponseOrder <- responseChanOrder
	}
	responseCheckTx := app.runCheckBCTxV3Concurrency(result)
	responseChanOrder := types3.ResponseChanOrder{
		Response: responseCheckTx,
		Index:    result.TxID,
	}
	//responsePool.Response <- responseCheckTx
	responsePool.ResponseOrder <- responseChanOrder
}

// ------------- add for support v1 transaction end ----------------
