package models

import (
	"git.solsynth.dev/hypernet/nexus/pkg/nex/cruda"
	"git.solsynth.dev/hypernet/wallet/pkg/proto"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	cruda.BaseModel

	Remark  string          `json:"remark"` // The usage of this transaction
	Amount  decimal.Decimal `json:"amount" type:"decimal(30,2);"`
	Payer   *Wallet         `json:"payer"`    // Who give the money
	Payee   *Wallet         `json:"payee"`    // Who get the money
	PayerID *uint           `json:"payer_id"` // Leave this field as nil means pay from the system
	PayeeID *uint           `json:"payee_id"` // Leave this field as nil means pay to the system
}

func (v *Transaction) ToTransactionInfo() *proto.TransactionInfo {
	amount, _ := v.Amount.Float64()
	info := &proto.TransactionInfo{
		Id:     uint64(v.ID),
		Amount: amount,
		Remark: v.Remark,
	}
	if v.PayerID != nil {
		info.PayerId = lo.ToPtr(uint64(*v.PayerID))
	}
	if v.PayeeID != nil {
		info.PayeeId = lo.ToPtr(uint64(*v.PayeeID))
	}
	return info
}
