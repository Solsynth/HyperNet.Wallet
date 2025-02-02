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
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

func getOrder(c *fiber.Ctx) error {
	orderId, _ := c.ParamsInt("orderId")

	var order models.Order
	if err := database.C.Where("id = ?", orderId).First(&order).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(order)
}

func createOrder(c *fiber.Ctx) error {
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

	order := models.Order{
		Status:   models.OrderStatusPending,
		Remark:   data.Remark,
		Amount:   decimal.NewFromFloat(data.Amount),
		ClientID: &client.ID,
	}

	// System client, spec payee was not allowed
	if client.AccountID != nil && data.PayeeID != nil {
		var payee models.Wallet
		if err := database.C.Where("id = ?", data.PayeeID).First(&payee).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("payee id %d not found", data.PayeeID))
		} else {
			order.Payee = &payee
			order.PayeeID = &payee.ID
		}
	}
	if data.PayerID != nil {
		var payer models.Wallet
		if err := database.C.Where("id = ?", data.PayerID).First(&payer).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("payer id %d not found", data.PayerID))
		} else {
			order.Payer = &payer
			order.PayerID = &payer.ID
		}
	}

	if err := database.C.Create(&order).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(order)
}

func payOrder(c *fiber.Ctx) error {
	if err := sec.EnsureAuthenticated(c); err != nil {
		return err
	}
	user := c.Locals("nex_user").(*sec.UserInfo)

	orderId, _ := c.ParamsInt("orderId")

	var data struct {
		WalletPassword string `json:"wallet_password" validate:"required"`
	}

	if err := exts.BindAndValidate(c, &data); err != nil {
		return err
	}

	var order models.Order
	if err := database.C.Where("id = ?", orderId).First(&order).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	var payer *models.Wallet
	if order.PayerID != nil {
		if err := database.C.Where("id = ?", order.PayerID).First(&payer).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "order payer wallet was not found")
		} else if payer.AccountID != user.ID {
			return fiber.NewError(fiber.StatusForbidden, "the order cannot paid by you")
		}
	} else {
		if err := database.C.Where("account_id = ?", order.ClientID).First(&payer).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "account wallet was not found")
		}
	}
	if payer != nil {
		if bcrypt.CompareHashAndPassword([]byte(payer.Password), []byte(data.WalletPassword)) != nil {
			return fiber.NewError(fiber.StatusForbidden, "invalid wallet password")
		}
	}

	var payee *models.Wallet
	if order.PayeeID != nil {
		if err := database.C.Where("id = ?", order.PayeeID).First(&payee).Error; err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "order payee wallet was not found")
		}
	}

	if tran, err := services.MakeTransaction(order.Amount.InexactFloat64(), order.Remark, payer, payee); err != nil {
		return fiber.NewError(fiber.StatusPaymentRequired, err.Error())
	} else {
		if err := database.C.Model(&order).Updates(&models.Order{
			Status:        models.OrderStatusPaid,
			TransactionID: &tran.ID,
		}).Error; err != nil {
			// Do refund
			_, _ = services.MakeTransaction(
				order.Amount.InexactFloat64(),
				fmt.Sprintf("%s - #%d Refund", order.Remark, order.ID),
				payee,
				payer,
			)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(order)
}
