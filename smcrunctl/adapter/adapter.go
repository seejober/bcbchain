package adapter

import (
	"github.com/bcbchain/bcbchain/burrow"
	"github.com/bcbchain/bcbchain/common/statedbhelper" //blacklist to del
	"github.com/bcbchain/bcbchain/smcrunctl/controllermgr"
	"github.com/bcbchain/bcbchain/smcrunctl/invokermgr"
	"github.com/bcbchain/bclib/types"
	types2 "github.com/bcbchain/bclib/tendermint/abci/types"
	"sync"

	"github.com/bcbchain/bclib/tendermint/tmlibs/log"
)

//Adapter objact of adapter
type Adapter struct {
	logger log.Logger
}

var (
	adpt *Adapter
	once sync.Once
)

//GetInstance get or create adapter instance
func GetInstance() *Adapter {
	once.Do(func() {
		adpt = &Adapter{}
	})
	return adpt
}

//Init init adapter before using it
//nolint errcheck
func (ad *Adapter) Init(log log.Logger, rpcPort int) {
	ad.logger = log
	controllermgr.GetInstance().Init(log, rpcPort)

	//Starting RPC server
	go start(rpcPort, log)
}

//Health get health status
func (ad *Adapter) Health() *types.Health {
	return controllermgr.GetInstance().Health()
}

//InvokeTx calls invokermgr's invoke function
func (ad *Adapter) InvokeTx(
	blockHeader types2.Header,
	transID, txID int64,
	sender types.Address,
	tx types.Transaction,
	publicKey types.PubKey,
	txHash types.Hash,
	blockHash types.Hash) *types.Response {
	// Sender can do nothing if it's in black list
	if statedbhelper.CheckBlackList(transID, txID, sender) == true {
		err := types.BcError{
			ErrorCode: types.ErrDealFailed,
		}
		return &types.Response{
			Code: err.ErrorCode,
			Log:  err.Error(),
		}
	}

	methodID := tx.Messages[len(tx.Messages)-1].MethodID
	if methodID == 0 || methodID == 0xFFFFFFFF {
		if !statedbhelper.CheckBVMEnable(transID, txID) {
			return &types.Response{
				Code: types.ErrLogicError,
				Log:  "BVM is disabled",
			}
		}
		return burrow.GetInstance(ad.logger).InvokeTx(blockHeader, blockHash, transID, txID, sender, tx, publicKey)
	}

	return invokermgr.GetInstance().InvokeTx(blockHeader, transID, txID, sender, tx, publicKey, txHash, blockHash)
}

//Commit commit transaction
func (ad *Adapter) Commit(transID int64) {
	invokermgr.GetInstance().Commit(transID)
}

//Rollback rollback transaction
func (ad *Adapter) Rollback(transID int64) {
	invokermgr.GetInstance().Rollback(transID)
}

//RollbackTx rollback transaction
func (ad *Adapter) RollbackTx(transID, txID int64) {
	invokermgr.GetInstance().RollbackTx(transID, txID)
}

// InitSMC init or upgrade chain for smart contact
func (ad *Adapter) InitOrUpdateSMC(transId, txId int64, header types2.Header, contractAddr, owner types.Address, isUpgarde bool) (result *types.Response) {
	result = invokermgr.GetInstance().InitOrUpdateSMC(transId, txId, header, contractAddr, owner, isUpgarde)
	return
}

// InitSMC mining for smart contact
func (ad *Adapter) Mine(transId, txId int64, header types2.Header, contractAddr, owner types.Address) (result *types.Response) {
	result = invokermgr.GetInstance().Mine(transId, txId, header, contractAddr, owner)
	return
}
