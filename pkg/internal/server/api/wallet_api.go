package api

import (
	"errors"
	"fmt"

	"git.solsynth.dev/hypernet/wallet/pkg/internal/server/exts"
	"golang.org/x/crypto/bcrypt"

	"git.solsynth.dev/hypernet/nexus/pkg/nex/sec"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func createWallet(c *fiber.Ctx) error {
	if err := sec.EnsureGrantedPerm(c, "CreateWallet", true); err != nil {
		return err
	}
	user := c.Locals("nex_user").(*sec.UserInfo)

	var data struct {
		Password string `json:"password" validate:"min=4"`
	}

	if err := exts.BindAndValidate(c, &data); err != nil {
		return err
	}

	var wallet models.Wallet
	if err := database.C.Where("account_id = ?", user.ID).
		First(&wallet).Error; err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusConflict, "wallet already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to hash your password: %v", err))
	}

	wallet = models.Wallet{
		Balance:   decimal.NewFromInt(0),
		Password:  string(hashed),
		AccountID: user.ID,
	}

	if err := database.C.Create(&wallet).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(wallet)
}

func getMyWallet(c *fiber.Ctx) error {
	if err := sec.EnsureAuthenticated(c); err != nil {
		return err
	}
	user := c.Locals("nex_user").(*sec.UserInfo)

	var wallet models.Wallet
	if err := database.C.Where("account_id = ?", user.ID).First(&wallet).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(wallet)
}
