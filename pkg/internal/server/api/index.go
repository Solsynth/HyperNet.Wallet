package api

import (
	"github.com/gofiber/fiber/v2"
)

func MapAPIs(app *fiber.App, baseURL string) {
	api := app.Group(baseURL).Name("API")
	{
		wallet := api.Group("/wallet").Name("Wallet API")
		{
			wallet.Get("/me", getMyWallet)
		}

		transaction := api.Group("/transaction").Name("Transaction API")
		{
			transaction.Get("/me", getTransaction)
		}
	}
}
