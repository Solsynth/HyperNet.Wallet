package models

import (
	"git.solsynth.dev/hypernet/nexus/pkg/nex/cruda"
	"git.solsynth.dev/hypernet/wallet/pkg/proto"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	cruda.BaseModel

	Balance   decimal.Decimal `json:"balance" sql:"type:decimal(30,2);"`
	Password  string          `json:"password"`
	AccountID uint            `json:"account_id"`
}

func (v *Wallet) ToWalletInfo() *proto.WalletInfo {
	balance, _ := v.Balance.Float64()
	return &proto.WalletInfo{
		Id:        uint64(v.ID),
		Balance:   balance,
		AccountId: uint64(v.AccountID),
	}
}
