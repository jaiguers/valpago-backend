package routes

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/usuario/valpago-backend/internal/db"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Lastname  string             `json:"lastname" bson:"lastname"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
	Phone     string             `json:"phone" bson:"phone"`
	Role      string             `json:"role" bson:"role"`
	IsActive  bool               `json:"isActive" bson:"isActive"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Lastname string `json:"lastname" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Phone    string `json:"phone" validate:"required"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	IsActive *bool  `json:"isActive"`
}

func createUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Check if user already exists
	var existingUser User
	err := db.Mongo().Collection("users").FindOne(c.Request().Context(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "User already exists"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "user" // Default role
	}

	// Create new user
	user := User{
		Name:      req.Name,
		Lastname:  req.Lastname,
		Email:     req.Email,
		Password:  string(hashedPassword), // Store hashed password
		Phone:     req.Phone,
		Role:      role,
		IsActive:  true, // New users are active by default
		CreatedAt: time.Now(),
	}

	result, err := db.Mongo().Collection("users").InsertOne(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	user.Password = "" // Don't return password

	return c.JSON(http.StatusCreated, user)
}

func listUsers(c echo.Context) error {
	cursor, err := db.Mongo().Collection("users").Find(c.Request().Context(), bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch users"})
	}
	defer cursor.Close(c.Request().Context())

	var users []User
	if err = cursor.All(c.Request().Context(), &users); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode users"})
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	return c.JSON(http.StatusOK, users)
}

func getUserByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var user User
	err = db.Mongo().Collection("users").FindOne(c.Request().Context(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}

	user.Password = "" // Don't return password
	return c.JSON(http.StatusOK, user)
}

func updateUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Lastname != "" {
		update["lastname"] = req.Lastname
	}
	if req.Email != "" {
		update["email"] = req.Email
	}
	if req.Phone != "" {
		update["phone"] = req.Phone
	}
	if req.Role != "" {
		update["role"] = req.Role
	}
	if req.IsActive != nil {
		update["isActive"] = *req.IsActive
	}

	if len(update) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No fields to update"})
	}

	result, err := db.Mongo().Collection("users").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User updated successfully"})
}
