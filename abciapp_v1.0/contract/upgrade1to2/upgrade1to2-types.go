package upgrade1to2

import (
	"github.com/bcbchain/bcbchain/abciapp_v1.0/contract/smcapi"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

type Upgrade1to2 struct {
	*smcapi.SmcApi
	GenesisOrg   std.Organization
	V2TokenIssue std.Contract
}

// Contract contract info
type Contract struct {
	Name       string         `json:"name,omitempty"`
	Version    string         `json:"version,omitempty"`
	CodeByte   types.HexBytes `json:"codeByte,omitempty"`
	CodeHash   string         `json:"codeHash,omitempty"`
	CodeDevSig Signature      `json:"codeDevSig,omitempty"`
	CodeOrgSig Signature      `json:"codeOrgSig,omitempty"`
}

// Signature sig for contract code
type Signature struct {
	PubKey    string `json:"pubkey"`
	Signature string `json:"signature"`
}
