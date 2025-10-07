package routes

import (
	"github.com/labstack/echo/v4"
)

func Register(e *echo.Echo) {
	// API routes
	api := e.Group("/api")

	// Users routes
	api.POST("/users", createUser)
	api.GET("/users", listUsers)
	api.GET("/users/:id", getUserByID)
	api.PUT("/users/:id", updateUser)

	// Auth routes
	api.POST("/auth/login", login)

	// Transactions routes
	api.POST("/transactions/create", createTransaction)
	api.GET("/transactions", listTransactions)
	api.PUT("/transactions/:id/status", updateTransactionStatus)
	api.PUT("/transactions/:id/review", reviewTransaction)
	api.PUT("/transactions/:id/approve", approveTransaction)
	api.PUT("/transactions/:id/reject", rejectTransaction)
}
