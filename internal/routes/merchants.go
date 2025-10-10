package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/usuario/valpago-backend/internal/db"
)

// Merchant representa un comercio
type Merchant struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Responsible string             `json:"responsible" bson:"responsible"`
	Name        string             `json:"name" bson:"name"`
	Phone       string             `json:"phone" bson:"phone"`
	Accounts    []string           `json:"accounts" bson:"accounts"`
}

// CreateMerchant crea un nuevo comercio
func CreateMerchant(c echo.Context) error {
	var merchant Merchant
	if err := c.Bind(&merchant); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validar campos requeridos (el ID lo genera el servidor)
	if merchant.Responsible == "" || merchant.Name == "" || merchant.Phone == "" || merchant.Accounts == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields: responsible, name, phone, accounts"})
	}

	// Insertar en MongoDB
	newMerchant, err := db.Mongo().Collection("merchants").InsertOne(c.Request().Context(), merchant)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create merchant"})
	}

	return c.JSON(http.StatusCreated, newMerchant)
}

// UpdateMerchant actualiza un comercio existente
func UpdateMerchant(c echo.Context) error {
	idStr := c.Param("id")
	merchantID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID"})
	}

	var merchant Merchant
	if err := c.Bind(&merchant); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validar campos requeridos
	if merchant.Responsible == "" || merchant.Name == "" || merchant.Phone == "" || merchant.Accounts == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields: responsible, name, phone, accounts"})
	}

	// Actualizar en MongoDB
	result, err := db.Mongo().Collection("merchants").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": merchantID},
		bson.M{"$set": bson.M{
			"responsible": merchant.Responsible,
			"name":        merchant.Name,
			"phone":       merchant.Phone,
			"accounts":    merchant.Accounts,
		}},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update merchant"})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Merchant not found"})
	}

	// Obtener el comercio actualizado
	var updatedMerchant Merchant
	err = db.Mongo().Collection("merchants").FindOne(c.Request().Context(), bson.M{"_id": merchantID}).Decode(&updatedMerchant)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve updated merchant"})
	}

	return c.JSON(http.StatusOK, updatedMerchant)
}

// ListMerchants obtiene todos los comercios
func ListMerchants(c echo.Context) error {
	// Obtener parámetros de paginación
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Calcular skip
	skip := (page - 1) * limit

	// Opciones de consulta
	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "name", Value: 1}}) // Ordenar por nombre

	// Buscar comercios
	cursor, err := db.Mongo().Collection("merchants").Find(c.Request().Context(), bson.M{}, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve merchants"})
	}
	defer cursor.Close(c.Request().Context())

	var merchants []Merchant
	if err = cursor.All(c.Request().Context(), &merchants); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode merchants"})
	}

	// Obtener total de comercios para paginación
	total, err := db.Mongo().Collection("merchants").CountDocuments(c.Request().Context(), bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count merchants"})
	}

	// Respuesta con paginación
	response := map[string]interface{}{
		"merchants": merchants,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	return c.JSON(http.StatusOK, response)
}

// GetMerchant obtiene un comercio específico por ID
func GetMerchant(c echo.Context) error {
	merchantID := c.Param("id")
	if merchantID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Merchant ID is required"})
	}

	var merchant Merchant
	err := db.Mongo().Collection("merchants").FindOne(c.Request().Context(), bson.M{"id": merchantID}).Decode(&merchant)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Merchant not found"})
	}

	return c.JSON(http.StatusOK, merchant)
}

// DeleteMerchant elimina un comercio
func DeleteMerchant(c echo.Context) error {
	merchantID := c.Param("id")
	if merchantID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Merchant ID is required"})
	}

	result, err := db.Mongo().Collection("merchants").DeleteOne(c.Request().Context(), bson.M{"id": merchantID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete merchant"})
	}

	if result.DeletedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Merchant not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Merchant deleted successfully"})
}
