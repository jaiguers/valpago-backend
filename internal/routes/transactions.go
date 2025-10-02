package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/usuario/valpago-backend/internal/db"
)

type Transaction struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PaymentMethod      string             `json:"payment_method" bson:"payment_method"`
	Amount             float64            `json:"amount" bson:"amount"`
	DestinationAccount string             `json:"destination_account" bson:"destination_account"`
	Reference          string             `json:"reference" bson:"reference"`
	SourceAccount      string             `json:"source_account" bson:"source_account"`
	Beneficiary        string             `json:"beneficiary" bson:"beneficiary"`
	WhatsappPhone      string             `json:"whatsapp_phone" bson:"whatsapp_phone"`
	Status             string             `json:"status" bson:"status"`
	SupportURL         string             `json:"support_url" bson:"support_url"`
	Date               string             `json:"date" bson:"date"`
	UserID             string             `json:"userId" bson:"userId"`
	CreatedAt          time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// Estructura para recibir datos en español
type CreateTransactionRequestSpanish struct {
	MetodoPago         string `json:"metodo_pago" validate:"required"`
	Monto              string `json:"monto" validate:"required"`
	CuentaConsignacion string `json:"cuenta_consignacion" validate:"required"`
	Referencia         string `json:"referencia" validate:"required"`
	CuentaOrigen       string `json:"cuenta_origen" validate:"required"`
	Beneficiario       string `json:"beneficiario" validate:"required"`
	TelWhatsappSend    string `json:"tel_whatsapp_send" validate:"required"`
	Estado             string `json:"estado" validate:"required"`
	URLSoport          string `json:"url_soporte" validate:"required"`
	Date               string `json:"date" validate:"required"`
}

// Estructura para almacenar en inglés (estructura interna)
type CreateTransactionRequest struct {
	UserID             string  `json:"userId" validate:"required"`
	PaymentMethod      string  `json:"payment_method" validate:"required"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	DestinationAccount string  `json:"destination_account" validate:"required"`
	Reference          string  `json:"reference" validate:"required"`
	SourceAccount      string  `json:"source_account" validate:"required"`
	Beneficiary        string  `json:"beneficiary" validate:"required"`
	WhatsappPhone      string  `json:"whatsapp_phone" validate:"required"`
	Status             string  `json:"status" validate:"required"`
	SupportURL         string  `json:"support_url" validate:"required"`
	Date               string  `json:"date" validate:"required"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=PENDING REVIEW APPROVED REJECTED"`
}

// Función para mapear de español a inglés
func mapSpanishToEnglish(spanish CreateTransactionRequestSpanish, userID string) (CreateTransactionRequest, error) {
	// Parsear monto de string a float64
	amount, err := strconv.ParseFloat(spanish.Monto, 64)
	if err != nil {
		return CreateTransactionRequest{}, fmt.Errorf("invalid amount format: %v", err)
	}

	return CreateTransactionRequest{
		UserID:             userID,
		PaymentMethod:      spanish.MetodoPago,
		Amount:             amount,
		DestinationAccount: spanish.CuentaConsignacion,
		Reference:          spanish.Referencia,
		SourceAccount:      spanish.CuentaOrigen,
		Beneficiary:        spanish.Beneficiario,
		WhatsappPhone:      spanish.TelWhatsappSend,
		Status:             spanish.Estado,
		SupportURL:         spanish.URLSoport,
		Date:               spanish.Date,
	}, nil
}

func createTransaction(c echo.Context) error {
	// Check API key
	apiKey := c.Request().Header.Get("x-api-key")
	if apiKey == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "API key required"})
	}

	// In production, validate API key against database
	// For now, we'll accept any non-empty API key

	// Recibir datos en español
	var spanishReq CreateTransactionRequestSpanish
	if err := c.Bind(&spanishReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Obtener userID del header o del token JWT
	userID := c.Request().Header.Get("user-id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID required in header"})
	}

	// Log para debug
	fmt.Printf("Received user-id: '%s'\n", userID)
	fmt.Printf("User-id length: %d\n", len(userID))

	// Verificar que el usuario existe
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		fmt.Printf("Error parsing ObjectID: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID format"})
	}

	var user User
	err = db.Mongo().Collection("users").FindOne(c.Request().Context(), bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("User not found with ID: %s\n", userID)
			// Vamos a verificar si hay usuarios en la base de datos
			cursor, _ := db.Mongo().Collection("users").Find(c.Request().Context(), bson.M{})
			var users []User
			cursor.All(c.Request().Context(), &users)
			fmt.Printf("Total users in database: %d\n", len(users))
			for i, u := range users {
				fmt.Printf("User %d: ID=%s, Email=%s\n", i+1, u.ID.Hex(), u.Email)
			}
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	fmt.Printf("User found: %s (%s)\n", user.Name, user.Email)

	// Mapear de español a inglés
	req, err := mapSpanishToEnglish(spanishReq, userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Create transaction
	transaction := Transaction{
		PaymentMethod:      req.PaymentMethod,
		Amount:             req.Amount,
		DestinationAccount: req.DestinationAccount,
		Reference:          req.Reference,
		SourceAccount:      req.SourceAccount,
		Beneficiary:        req.Beneficiary,
		WhatsappPhone:      req.WhatsappPhone,
		Status:             req.Status,
		SupportURL:         req.SupportURL,
		Date:               req.Date,
		UserID:             req.UserID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	result, err := db.Mongo().Collection("transactions").InsertOne(c.Request().Context(), transaction)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction"})
	}

	transaction.ID = result.InsertedID.(primitive.ObjectID)

	// Send to Redis stream for processing
	if db.Rdb != nil {
		streamData := map[string]interface{}{
			"transaction_id": transaction.ID.Hex(),
			"user_id":        transaction.UserID,
			"amount":         transaction.Amount,
			"status":         transaction.Status,
			"timestamp":      time.Now().Unix(),
		}
		db.Rdb.XAdd(c.Request().Context(), &redis.XAddArgs{
			Stream: "valpago:transactions",
			Values: streamData,
		})
	}

	return c.JSON(http.StatusCreated, transaction)
}

func listTransactions(c echo.Context) error {
	cursor, err := db.Mongo().Collection("transactions").Find(c.Request().Context(), bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch transactions"})
	}
	defer cursor.Close(c.Request().Context())

	var transactions []Transaction
	if err = cursor.All(c.Request().Context(), &transactions); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode transactions"})
	}

	return c.JSON(http.StatusOK, transactions)
}

func updateTransactionStatus(c echo.Context) error {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
	}

	var req UpdateStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Update transaction status
	result, err := db.Mongo().Collection("transactions").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":     req.Status,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction"})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found"})
	}

	// Send notification via Redis stream
	if db.Rdb != nil {
		streamData := map[string]interface{}{
			"transaction_id": id.Hex(),
			"status":         req.Status,
			"timestamp":      time.Now().Unix(),
		}
		db.Rdb.XAdd(c.Request().Context(), &redis.XAddArgs{
			Stream: "valpago:notifications",
			Values: streamData,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Transaction status updated successfully"})
}
