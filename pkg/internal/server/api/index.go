package api

import (
	"github.com/gofiber/fiber/v2"
)

func MapAPIs(app *fiber.App, baseURL string) {
	api := app.Group(baseURL).Name("API")
	{
		wallet := api.Group("/wallets").Name("Wallet API")
		{
			wallet.Post("/me", createWallet)
			wallet.Get("/me", getMyWallet)
		}

		transaction := api.Group("/transactions").Name("Transaction API")
		{
			transaction.Get("/me", getTransaction)
			transaction.Get("/:id", getTransactionByID)
			transaction.Post("/", makeTransaction)
		}

		order := api.Group("/orders").Name("Order API")
		{
			order.Get("/:orderId", getOrder)
			order.Post("/:orderId/pay", payOrder)
			order.Post("/:orderId/cancel", cancelOrder)
			order.Post("/", createOrder)
		}
	}
}
