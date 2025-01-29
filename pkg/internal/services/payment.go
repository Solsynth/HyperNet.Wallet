package services

import (
	"fmt"

	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/models"
	"github.com/shopspring/decimal"
)

func MakeTransaction(amount float64, remark string, payer, payee *models.Wallet) (models.Transaction, error) {
	transaction := models.Transaction{
		Amount: decimal.NewFromFloat(amount),
		Remark: remark,
	}
	if payer != nil {
		transaction.PayerID = &payer.ID
	}
	if payee != nil {
		transaction.PayeeID = &payee.ID
	}

	tx := database.C.Begin()

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return transaction, err
	}

	if payer != nil {
		payer.Balance = payer.Balance.Sub(transaction.Amount)
		if err := tx.Model(payer).
			Updates(&models.Wallet{Balance: payer.Balance}).Error; err != nil {
			tx.Rollback()
			return transaction, fmt.Errorf("failed to update payer wallet balance: %w", err)
		}
	}
	if payee != nil {
		payee.Balance = payee.Balance.Add(transaction.Amount)
		if err := tx.Model(payee).
			Updates(&models.Wallet{Balance: payee.Balance}).Error; err != nil {
			tx.Rollback()
			return transaction, fmt.Errorf("failed to update payee wallet balance: %w", err)
		}
	}

	tx.Commit()

	return transaction, nil
}
