package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/madhouselabs/anybase/internal/user"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserHandler struct {
	userRepo user.Repository
}

func NewUserHandler(userRepo user.Repository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

// ListUsers returns all users (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get pagination parameters
	page := 1
	limit := 20
	
	// Calculate skip
	skip := (page - 1) * limit
	
	// Find users using the List method
	users, err := h.userRepo.List(ctx, bson.M{}, &options.FindOptions{
		Skip:  int64Ptr(int64(skip)),
		Limit: int64Ptr(int64(limit)),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}
	
	// Remove sensitive data
	for i := range users {
		users[i].Password = ""
	}
	
	// Get total count
	total, _ := h.userRepo.Count(ctx, bson.M{})
	
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetUser returns a specific user
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	ctx := c.Request.Context()
	
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	user, err := h.userRepo.GetByID(ctx, objID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	
	// Remove sensitive data
	user.Password = ""
	
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	
	var createData struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=6"`
		Role      string `json:"role" binding:"required,oneof=admin developer"`
		Active    bool   `json:"active"`
	}
	
	if err := c.ShouldBindJSON(&createData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}
	
	// Check if user with email already exists
	existingUser, _ := h.userRepo.GetByEmail(ctx, createData.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User with this email already exists",
		})
		return
	}
	
	// Hash password
	hashedPassword, err := user.HashPassword(createData.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}
	
	// Create new user
	newUser := &models.User{
		FirstName: createData.FirstName,
		LastName:  createData.LastName,
		Email:     createData.Email,
		Password:  hashedPassword,
		Role:      createData.Role,
		Active:    createData.Active,
		UserType:  "regular",
	}
	
	// Create user
	if err := h.userRepo.Create(ctx, newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}
	
	// Remove sensitive data
	newUser.Password = ""
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    newUser,
	})
}

// UpdateUser updates a user's information
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	ctx := c.Request.Context()
	
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	var updateData struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		Active    bool   `json:"active"`
	}
	
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	
	// Build update map
	updateMap := bson.M{}
	if updateData.FirstName != "" {
		updateMap["first_name"] = updateData.FirstName
	}
	if updateData.LastName != "" {
		updateMap["last_name"] = updateData.LastName
	}
	if updateData.Email != "" {
		updateMap["email"] = updateData.Email
	}
	if updateData.Role != "" {
		updateMap["role"] = updateData.Role
	}
	updateMap["active"] = updateData.Active
	
	// Update user directly with map
	if err := h.userRepo.UpdateRaw(ctx, objID, updateMap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}
	
	// Get updated user
	updatedUser, err := h.userRepo.GetByID(ctx, objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch updated user",
		})
		return
	}
	
	// Remove sensitive data
	updatedUser.Password = ""
	
	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    updatedUser,
	})
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	ctx := c.Request.Context()
	
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	// Don't allow deleting the requesting user
	requestingUserID := c.GetString("userID")
	if requestingUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete your own account",
		})
		return
	}
	
	if err := h.userRepo.Delete(ctx, objID); err != nil {
		if err == user.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// GetUserProfile returns the current user's profile
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID := c.GetString("userID")
	ctx := c.Request.Context()
	
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	user, err := h.userRepo.GetByID(ctx, objID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	
	// Remove sensitive data
	user.Password = ""
	
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateUserProfile updates the current user's profile
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	userID := c.GetString("userID")
	ctx := c.Request.Context()
	
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	var updateData struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	
	// Build update struct
	userUpdate := &models.UserUpdate{
		FirstName: updateData.FirstName,
		LastName:  updateData.LastName,
	}
	
	// Update user
	if err := h.userRepo.Update(ctx, objID, userUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile",
		})
		return
	}
	
	// Get updated user
	updatedUser, err := h.userRepo.GetByID(ctx, objID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch updated profile",
		})
		return
	}
	
	// Remove sensitive data
	updatedUser.Password = ""
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    updatedUser,
	})
}