package api

import (
	"fmt"

	"git.solsynth.dev/hypernet/nexus/pkg/nex/sec"
	"git.solsynth.dev/hypernet/passport/pkg/authkit"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/gap"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/models"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/server/exts"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/services"
	"github.com/gofiber/fiber/v2"
)

func getTransaction(c *fiber.Ctx) error {
	take := c.QueryInt("take", 0)
	offset := c.QueryInt("offset", 0)

	if err := sec.EnsureAuthenticated(c); err != nil {
		return err
	}
	user := c.Locals("nex_user").(*sec.UserInfo)

	var wallet models.Wallet
	if err := database.C.Where("account_id = ?", user.ID).First(&wallet).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	var count int64
	if err := database.C.Model(&models.Transaction{}).Where("payer_id = ? OR payee_id = ?", user.ID, user.ID).
		Count(&count).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var transactions []models.Transaction
	if err := database.C.Where("payer_id = ? OR payee_id = ?", user.ID, user.ID).
		Limit(take).Offset(offset).
		Order("created_at DESC").
		Find(&transactions).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"count": count,
		"data":  transactions,
	})
}

func getTransactionByID(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id", 0)

	if err := sec.EnsureAuthenticated(c); err != nil {
		return err
	}
	user := c.Locals("nex_user").(*sec.UserInfo)

	var transaction models.Transaction
	if err := database.C.Where("id = ?", id).
		Preload("Payer").Preload("Payee").
		First(&transaction).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	if transaction.Payer.AccountID != user.ID && transaction.Payee.AccountID != user.ID {
		return fiber.NewError(fiber.StatusForbidden, "you are not related to this transaction")
	}

	return c.JSON(transaction)
}

func makeTransaction(c *fiber.Ctx) error {
	var data struct {
		ClientID     string  `json:"client_id" validate:"required"`
		ClientSecret string  `json:"client_secret" validate:"required"`
		Remark       string  `json:"remark" validate:"required"`
		Amount       float64 `json:"amount" validate:"required"`
		PayeeID      *uint   `json:"payee_id"`
		PayerID      *uint   `json:"payer_id"`
	}

	if err := exts.BindAndValidate(c, &data); err != nil {
		return err
	}

	// Validating client
	client, err := authkit.GetThirdClientByAlias(gap.Nx, data.ClientID, &data.ClientSecret)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("could not get client info: %v", err))
	}

	// System client, spec payer was not allowed
	var payee, payer *models.Wallet
	if client.AccountID != nil && data.PayeeID != nil {
		if err := database.C.Where("id = ? AND account_id = ?", data.PayerID, client.AccountID).First(&payer).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("payer id %d not found", data.PayerID))
		}
	}
	if client.AccountID != nil && payer == nil {
		return fiber.NewError(fiber.StatusBadRequest, "payer is required if issuer is individual")
	}

	if data.PayeeID != nil {
		if err := database.C.Where("id = ?", data.PayeeID).First(&payee).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("payer id %d not found", data.PayeeID))
		}
	}

	if payee == nil && payer == nil {
		return fiber.NewError(fiber.StatusBadRequest, "payee and payer cannot be both blank")
	}

	tran, err := services.MakeTransaction(data.Amount, data.Remark, payer, payee)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(tran)
}
