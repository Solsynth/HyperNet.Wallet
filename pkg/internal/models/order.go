package models

import (
	"git.solsynth.dev/hypernet/nexus/pkg/nex/cruda"
	"github.com/shopspring/decimal"
)

const (
	OrderStatusPending = iota
	OrderStatusPaid
	OrderStatusCanceled
)

type Order struct {
	cruda.BaseModel

	Status        int             `json:"status"`
	Remark        string          `json:"remark"`
	Amount        decimal.Decimal `json:"amount"`
	Payer         *Wallet         `json:"payer"`
	Payee         *Wallet         `json:"payee"`
	PayerID       *uint           `json:"payer_id"`
	PayeeID       *uint           `json:"payee_id"`
	Transaction   *Transaction    `json:"transaction"`
	TransactionID *uint           `json:"transaction_id"`
	ClientID      *uint           `json:"client_id"`
}
