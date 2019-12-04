package statedb

import (
	"bytes"
	"common/bcdb"
	"encoding/json"
	"math/big"
	"sort"
	"strings"

	"blockchain/abciapp_v1.0/smc"
	"blockchain/abciapp_v1.0/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
)

type StateDB struct {
	DB          *bcdb.GILevelDB
	BlockBuffer map[string][]byte
}

func NewStateDB(db *bcdb.GILevelDB) *StateDB {
	return &StateDB{db, nil}
}

func (sdb *StateDB) GetGenesisToken() (*types.IssueToken, error) {
	key := keyOfGenesisToken()
	tokenData, err := sdb.Get(key)
	if err != nil {
		return nil, err
	}
	if tokenData == nil {
		return nil, nil
	}
	var token types.IssueToken
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (sdb *StateDB) getChildKeys(key string) ([]string, error) {

	value, err := sdb.Get(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, err
	}

	var strChildKeys []string
	err = json.Unmarshal(value, &strChildKeys)
	if err != nil {
		return nil, err
	}
	return strChildKeys, nil
}

//先取出ChildKeys，然后判断新加的childKey是否在列表中，如果在，直接返回，如果不在，增加然后保存回去。
//特别注意：调用该函数之前不能通过batch修改key对应的数据，
// 因为：该函数中会取key对应的子key列表，如果在调用此函数之前，通过batch修改了子key列表，函数中会把数据冲掉，造成错误。
func (sdb *StateDB) addChildKey(key string, childKey string) error {
	childKeys, err := sdb.getChildKeys(key)
	if err != nil {
		return err
	}

	index := sort.SearchStrings(childKeys, childKey)
	if index == len(childKeys) { // insert tail
		childKeys = append(childKeys, childKey)
	} else if childKeys[index] == childKey { //equal
		return nil
	} else {
		childKeys = append(childKeys[:index], append([]string{childKey}, childKeys[index:]...)...)
	}

	childKeysData, err := json.Marshal(childKeys)
	if err != nil {
		return err
	}
	sdb.Set(key, []byte(childKeysData))

	return nil
}

func (sdb *StateDB) addChildKeyEx(key string, childKey string) (map[string][]byte, error) {
	childKeys, err := sdb.getChildKeys(key)
	if err != nil {
		return nil, err
	}

	index := sort.SearchStrings(childKeys, childKey)
	if index == len(childKeys) { // insert tail
		childKeys = append(childKeys, childKey)
	} else if childKeys[index] == childKey { //equal
		return nil, nil
	} else {
		childKeys = append(childKeys[:index], append([]string{childKey}, childKeys[index:]...)...)
	}

	childKeysData, err := json.Marshal(childKeys)
	if err != nil {
		return nil, err
	}
	sdb.Set(key, []byte(childKeysData))

	return map[string][]byte{key: []byte(childKeysData)}, nil
}

//先设置合约，再设置代币
func (sdb *StateDB) SetGenesisToken(genesisToken *types.IssueToken) error {
	key := keyOfGenesisToken()
	genesisTokenData, err := json.Marshal(genesisToken)
	if err != nil {
		return err
	}
	//保存到创世路径
	sdb.Set(key, genesisTokenData)

	//保存到token路径
	key = keyOfToken(genesisToken.Address)
	sdb.Set(key, genesisTokenData)
	sdb.addChildKey(keyOfTokenAll(), genesisToken.Address)

	addressData, err := json.Marshal(genesisToken.Address)
	if err != nil {
		return err
	}
	sdb.Set(keyOfTokenName(genesisToken.Name), addressData)
	sdb.Set(keyOfTokenSymbol(genesisToken.Symbol), addressData)
	baseGasPriceData, err := json.Marshal(&genesisToken.GasPrice)
	sdb.Set(keyOfTokenBaseGasPrice(), baseGasPriceData)

	//设置外部账户的余额
	token := types.TokenBalance{genesisToken.Address, genesisToken.TotalSupply}
	tokenData, err := json.Marshal(&token)
	if err != nil {
		return err
	}
	key = keyOfAccount(genesisToken.Owner)
	childkey := keyOfAccountToken(genesisToken.Owner, genesisToken.Address)
	sdb.Set(childkey, tokenData)
	return sdb.addChildKey(key, childkey)
}

func (sdb *StateDB) GetGenesisContract(contractAddr smc.Address) (*types.Contract, error) {
	key := keyOfGenesisContract(contractAddr)
	contractData, err := sdb.Get(key)
	if err != nil {
		return nil, err
	}
	if contractData == nil {
		return nil, nil
	}
	var contract types.Contract
	err = json.Unmarshal(contractData, &contract)
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (sdb *StateDB) SetGenesisContract(contract *types.Contract) error {
	//检查合约是否存在，智能合约只允许写一次
	con, err := sdb.GetContract(contract.Address)
	if err != nil {
		return err
	}
	if con != nil {
		return errors.New("Repeated calls to SetGenesisContract()")
	}

	contractData, err := json.Marshal(contract)
	if err != nil {
		return err
	}

	//设置创世合约信息
	key := keyOfGenesisContract(contract.Address)
	sdb.Set(key, contractData)
	//增加到创世合约列表
	sdb.addChildKey(keyOfGenesisContracts(), contract.Address)

	//设置合约账户信息
	key = keyOfContract(contract.Address)
	sdb.Set(key, contractData)
	sdb.addChildKey(keyOfContractAll(), contract.Address)

	//保存智能合约Owner的相关信息
	key = keyOfAccount(contract.Owner)
	childKey := keyOfAccountContracts(contract.Owner)
	err = sdb.addChildKey(childKey, contract.Address)
	if err != nil {
		return err
	}
	err = sdb.addChildKey(key, childKey)
	if err != nil {
		return err
	}

	return nil
}

func (sdb *StateDB) GetGenesisContractList() ([]string, error) {
	key := keyOfGenesisContracts()
	strContracts, err := sdb.getChildKeys(key)
	return strContracts, err
}

func (sdb *StateDB) GetContractAddrList() ([]string, error) {
	key := keyOfContractAll()
	strContracts, err := sdb.getChildKeys(key)
	return strContracts, err
}

func (sdb *StateDB) GetTokenAddrList() ([]string, error) {
	key := keyOfTokenAll()
	strContracts, err := sdb.getChildKeys(key)
	return strContracts, err
}

func (sdb *StateDB) GetTokenBalListWithAccAddr(accountAddr smc.Address) ([]types.TokenBalance, error) {
	var tokenBalList []types.TokenBalance
	value, err := sdb.Get(keyOfAccount(accountAddr))
	if err != nil {
		return tokenBalList, err
	}

	if len(value) == 0 {
		return tokenBalList, nil
	}

	var childKeys []smc.Address
	err = json.Unmarshal(value, &childKeys)
	if err != nil {
		return tokenBalList, err
	}

	for _, childKey := range childKeys {
		if strings.HasPrefix(childKey, "/account/ex/") && strings.Contains(childKey, "token") {
			value, err = sdb.Get(childKey)
			if err == nil {
				var tokenBal types.TokenBalance
				json.Unmarshal(value, &tokenBal)
				tokenBalList = append(tokenBalList, tokenBal)
			}
		}
	}

	return tokenBalList, nil
}

func (sdb *StateDB) GetValidator(nodeAddr string) (*types.Validator, error) {
	value, err := sdb.Get(keyOfValidator(nodeAddr))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	var validator types.Validator
	err = json.Unmarshal(value, &validator)
	if err != nil {
		return nil, err
	}
	return &validator, nil
}

func (sdb *StateDB) GetAllValidators() ([]types.Validator, error) {
	nodeAddrs, err := sdb.getChildKeys(keyOfValidators())
	if err != nil {
		return nil, err
	}

	var validators = make([]types.Validator, 0)
	for _, nodeAddr := range nodeAddrs {
		val, err := sdb.GetValidator(nodeAddr)
		if err != nil {
			return nil, err
		}
		validators = append(validators, *val)
	}
	return validators, nil
}

func (sdb *StateDB) GetWorldAppState() (*abci.AppState, error) {
	key := keyOfWorldAppState()
	value, err := sdb.Get(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	var appState abci.AppState
	err = json.Unmarshal(value, &appState)
	if err != nil {
		return nil, err
	}
	return &appState, nil
}

func (sdb *StateDB) GetChainID() string {
	value, err := sdb.Get(keyOfGenesisChainId())
	if err != nil {
		panic(err)
	}
	if value == nil {
		panic(errors.New("ChainID is null"))
	}

	return string(value)
}

func (sdb *StateDB) SetStrategys(strategys []types.RewardStrategy) error {
	key := keyOfRewardStrategys()
	strategysData, err := json.Marshal(strategys)
	if err != nil {
		return err
	}
	sdb.Set(key, strategysData)
	return nil
}

func (sdb *StateDB) SetChainID(chainID string) {
	key := keyOfGenesisChainId()
	sdb.Set(key, []byte(chainID))
}

func (sdb *StateDB) SetWorldAppState(appState *abci.AppState) error {
	key := keyOfWorldAppState()
	appStateData, err := json.Marshal(appState)
	if err != nil {
		return err
	}
	sdb.Set(key, appStateData)
	return nil
}

func (sdb *StateDB) SetValidator(validator *types.Validator) error {
	key := keyOfValidator(validator.NodeAddr)
	validatorData, err := json.Marshal(validator)
	if err != nil {
		return err
	}
	sdb.Set(key, validatorData)
	return sdb.addChildKey(keyOfValidators(), validator.NodeAddr)
}

//先在内存中找，如果找不到，再到db中找
func (sdb *StateDB) GetBalance(contractAddress smc.Address, exAddress smc.Address) (big.Int, error) {
	key := keyOfAccountToken(exAddress, contractAddress)
	tokenData, err := sdb.Get(key)
	if err != nil {
		return *big.NewInt(0), err
	}
	if tokenData == nil {
		return *big.NewInt(0), nil
	}

	var token types.TokenBalance
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		return *big.NewInt(0), err
	}
	return token.Balance, nil
}

//Get nonce of the account
func (sdb *StateDB) GetAccountNonce(exAddress smc.Address) (uint64, error) {
	var lastNonce uint64
	key := keyOfAccountNonce(exAddress)
	accountData, err := sdb.Get(key)
	if err != nil {
		return 0, err
	}
	if accountData == nil {
		lastNonce = 0
	} else {
		var account types.AccountInfo
		err = json.Unmarshal(accountData, &account)
		if err != nil {
			return 0, err
		}
		lastNonce = account.Nonce
	}
	return lastNonce, nil
}

//CheckTx需要调用此接口检查nonce
func (sdb *StateDB) CheckAccountNonce(exAddress smc.Address, nonce uint64) error {
	//根据合约地址，账户地址，构造出key，然后保存
	var lastNonce uint64
	key := keyOfAccountNonce(exAddress)
	accountData, err := sdb.Get(key)
	if err != nil {
		return err
	}
	if accountData == nil {
		lastNonce = 0
	} else {
		var account types.AccountInfo
		err = json.Unmarshal(accountData, &account)
		if err != nil {
			return err
		}
		lastNonce = account.Nonce
	}

	if (lastNonce + 1) != nonce {
		return errors.New("Address:" + exAddress + " nonce invalid!")
	}
	return nil
}

//DeliverTx需要调用此接口检查并设置nonce
func (sdb *StateDB) SetAccountNonce(exAddress smc.Address, nonce uint64) (nonceBuffer map[string][]byte, err error) {
	//根据合约地址，账户地址，构造出key，然后保存
	err = sdb.CheckAccountNonce(exAddress, nonce)
	if err != nil {
		return
	}

	account := types.AccountInfo{Nonce: nonce}
	accountData, err := json.Marshal(&account)
	if err != nil {
		return
	}
	key := keyOfAccount(exAddress)
	childKey := keyOfAccountNonce(exAddress)
	data, err := sdb.Get(childKey)
	if err != nil {
		return
	}
	nonceBuffer = make(map[string][]byte)
	if data == nil {
		var tmpBuffer map[string][]byte
		tmpBuffer, err = sdb.addChildKeyEx(key, childKey)
		if err != nil {
			return
		}
		nonceBuffer = tmpBuffer
	}

	nonceBuffer[childKey] = []byte(accountData)
	sdb.Set(childKey, []byte(accountData))
	return
}

func (sdb *StateDB) GetToken(contractAddr smc.Address) (*types.IssueToken, error) {
	key := keyOfToken(contractAddr)
	tokenData, err := sdb.Get(key)
	if err != nil {
		return nil, err
	}
	if tokenData == nil {
		return nil, nil
	}

	var token types.IssueToken
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (sdb *StateDB) GetContract(contractAddr smc.Address) (*types.Contract, error) {
	key := keyOfContract(contractAddr)
	contractData, err := sdb.Get(key)
	if err != nil {
		return nil, err
	}
	if contractData == nil {
		return nil, nil
	}
	var contract types.Contract
	err = json.Unmarshal(contractData, &contract)
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (sdb *StateDB) Get(key string) ([]byte, error) {
	v, ok := sdb.BlockBuffer[key]
	if ok {
		return v, nil
	}
	return sdb.DB.Get([]byte(key))
}

func (sdb *StateDB) Set(key string, value []byte) {
	if sdb.BlockBuffer == nil {
		panic("BlockBuffer is nil")
	}
	sdb.BlockBuffer[key] = value
}

func (sdb *StateDB) NewTxState(contractAddress smc.Address, senderAddress smc.Address) *TxState {
	return &TxState{sdb, contractAddress, senderAddress, nil}
}

func (sdb *StateDB) BeginBlock() {
	//创建二级缓存
	sdb.BlockBuffer = make(map[string][]byte)
}

func (sdb *StateDB) EndBlock() {
	//
}

func (sdb *StateDB) CommitTx(txBuffer map[string][]byte) {
	//多笔交易奖励gas时，余额是从blockBuffer中获取到的，所以可以直接覆盖
	for k, v := range txBuffer {
		sdb.BlockBuffer[k] = v
	}
}

func (sdb *StateDB) GetBlockBuffer() []byte {
	var keys []string
	//先遍历map把key加入到字符串数组中，然后对字符串数组进行排序
	for k := range sdb.BlockBuffer {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	//遍历字符串数组
	for _, k := range keys {
		v := sdb.BlockBuffer[k]
		buf.Write([]byte(k))
		buf.Write(v)
	}

	return buf.Bytes()
}

func (sdb *StateDB) CommitBlock() error {

	//创建batch，通过batch，操作数据库，清除Block缓存。
	batch := sdb.DB.NewBatch()

	//遍历BlockBuffer，batch.set
	for k, v := range sdb.BlockBuffer {
		batch.Set([]byte(k), v)
	}
	err := batch.Commit()
	if err != nil {
		panic(err)
	}
	sdb.BlockBuffer = nil //删除Block缓存
	return nil
}

func (sdb *StateDB) CommitBlockEx() {

	sdb.BlockBuffer = nil //删除Block缓存
}

func (sdb *StateDB) RollBlock() {
	//清除Block缓存
	sdb.BlockBuffer = nil
}
