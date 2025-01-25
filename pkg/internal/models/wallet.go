package models

import (
	"git.solsynth.dev/hypernet/nexus/pkg/nex/cruda"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	cruda.BaseModel

	Transactions []Transaction   `json:"transactions"`
	Balance      decimal.Decimal `json:"amount" sql:"type:decimal(30,2);"`
	AccountID    uint            `json:"account_id"`
}
