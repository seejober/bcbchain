package smcbuilder

import (
	"github.com/bcbchain/bclib/algorithm"
	"github.com/bcbchain/bcbchain/smcdocker"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/bcbchain/statedb"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bcbchain/bclib/tendermint/tmlibs/log"
)

func TestNewBuilder(t *testing.T) {

}

func TestBuilder_GetContractDllPath(t *testing.T) {
	logger := log.NewOldTMLogger(os.Stdout)
	statedb.Init("testdb", "127.0.0.1", "37888")
	statedb.NewTransaction()
	contractAddr := "contractAddress"
	orgID := "orgBtjfCSPCAJ84uQWcpNr74NLMWYm5SXzer"

	org := std.Organization{
		OrgID:            orgID,
		Name:             "org",
		OrgOwner:         "orgowner",
		ContractAddrList: []string{contractAddr},
		OrgCodeHash:      algorithm.CalcCodeHash("helloA"),
		Signers:          nil,
	}
	res, _ := jsoniter.Marshal(org)
	statedb.Set(1, 1, "/organization/"+orgID, res)

	codePath := "/Users/test/today/mydice2win.tar.gz"

	codeBytes, _ := ioutil.ReadFile(codePath)
	meta := std.ContractMeta{
		Name:         "mydice2win",
		ContractAddr: contractAddr,
		OrgID:        orgID,
		Version:      "1.0",
		EffectHeight: 10,
		LoseHeight:   0,
		CodeData:     codeBytes,
		CodeHash:     []byte("helloA"),
		CodeDevSig:   nil,
		CodeOrgSig:   nil,
	}
	resCode, _ := jsoniter.Marshal(meta)
	statedb.Set(1, 1, "/contract/code/"+contractAddr, resCode)

	con := std.Contract{
		Address:      contractAddr,
		Account:      "",
		Owner:        "",
		Name:         "",
		Version:      "1.0",
		CodeHash:     []byte("helloA"),
		EffectHeight: 10,
		LoseHeight:   0,
		KeyPrefix:    "",
		Methods:      nil,
		Interfaces:   nil,
		Token:        "",
		OrgID:        orgID,
	}
	resCon, _ := jsoniter.Marshal(con)
	statedb.Set(1, 1, "/contract/"+contractAddr, resCon)

	Init(logger, "/Users/test/test-bcchain")
	//p := b.GetContractDllPath(1, 1, orgID)
	//fmt.Println("RESULT:" + p)

	d := smcdocker.SMCDocker{}
	d.Init(&logger, "127.0.0.1:33998")
	d.GetContractInvokeUrl(1, 1, "contractAddr")
}

func TestBuilder_BuildContract(t *testing.T) {

}
