package services

import (
	"fmt"
	"math"

	"git.solsynth.dev/hypernet/passport/pkg/authkit"
	"git.solsynth.dev/hypernet/pusher/pkg/pushkit"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/gap"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/models"
	"github.com/shopspring/decimal"
)

func MakeTransaction(amount float64, remark, currency string, payer, payee *models.Wallet) (models.Transaction, error) {
	// Round amount to keep 2 decimal places
	amount = math.Round(amount*100) / 100

	transaction := models.Transaction{
		Amount:   decimal.NewFromFloat(amount),
		Remark:   remark,
		Currency: currency,
	}
	if payer != nil {
		isInsufficient := false
		switch currency {
		case "golden":
			isInsufficient = payer.GoldenBalance.LessThan(transaction.Amount)
		default:
			isInsufficient = payer.Balance.LessThan(transaction.Amount)
		}
		if isInsufficient {
			return transaction, fmt.Errorf("payer account has insufficient balance to pay this transaction")
		}
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
		switch currency {
		case "golden":
			payer.GoldenBalance = payer.GoldenBalance.Sub(transaction.Amount)
		default:
			payer.Balance = payer.Balance.Sub(transaction.Amount)
		}
		if err := tx.Model(payer).
			Updates(&models.Wallet{Balance: payer.Balance, GoldenBalance: payer.GoldenBalance}).Error; err != nil {
			tx.Rollback()
			return transaction, fmt.Errorf("failed to update payer wallet balance: %w", err)
		}
	}
	if payee != nil {
		switch currency {
		case "golden":
			payee.GoldenBalance = payee.GoldenBalance.Add(transaction.Amount)
		default:
			payee.Balance = payee.Balance.Add(transaction.Amount)
		}
		if err := tx.Model(payee).
			Updates(&models.Wallet{Balance: payee.Balance, GoldenBalance: payee.GoldenBalance}).Error; err != nil {
			tx.Rollback()
			return transaction, fmt.Errorf("failed to update payee wallet balance: %w", err)
		}
	}

	tx.Commit()

	if payer != nil {
		authkit.NotifyUser(gap.Nx, uint64(payer.AccountID), pushkit.Notification{
			Topic:    "wallet.transaction.new",
			Title:    fmt.Sprintf("Receipt #%d", transaction.ID),
			Subtitle: transaction.Remark,
			Body:     fmt.Sprintf("%.2f SRC removed from your wallet. Your new balance is %.2f", amount, payer.Balance.InexactFloat64()),
			Metadata: map[string]any{
				"id":     transaction.ID,
				"amount": amount,
				"balance": (func() float64 {
					switch currency {
					case "golden":
						return payer.GoldenBalance.InexactFloat64()
					default:
						return payer.Balance.InexactFloat64()
					}
				})(),
				"remark":   transaction.Remark,
				"currency": currency,
			},
			Priority: 0,
		})
	}
	if payee != nil {
		authkit.NotifyUser(gap.Nx, uint64(payee.AccountID), pushkit.Notification{
			Topic:    "wallet.transaction.new",
			Title:    fmt.Sprintf("Receipt #%d", transaction.ID),
			Subtitle: transaction.Remark,
			Body:     fmt.Sprintf("%.2f SRC added to your wallet. Your new balance is %.2f", amount, payee.Balance.InexactFloat64()),
			Metadata: map[string]any{
				"id":     transaction.ID,
				"amount": amount,
				"balance": (func() float64 {
					switch currency {
					case "golden":
						return payee.GoldenBalance.InexactFloat64()
					default:
						return payee.Balance.InexactFloat64()
					}
				})(),
				"remark":   transaction.Remark,
				"currency": currency,
			},
			Priority: 0,
		})
	}

	return transaction, nil
}
