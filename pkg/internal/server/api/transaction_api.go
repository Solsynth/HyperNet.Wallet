package api

import (
	"git.solsynth.dev/hypernet/nexus/pkg/nex/sec"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/models"
	"github.com/gofiber/fiber/v2"
)

func getTransaction(c *fiber.Ctx) error {
	take := c.QueryInt("take", 0)
	offset := c.QueryInt("offset", 0)

	if err := sec.EnsureAuthenticated(c); err != nil {
		return err
	}
	user := c.Locals("user").(*sec.UserInfo)

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
		Find(&transactions).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"count": count,
		"data":  transactions,
	})
}
