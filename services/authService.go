package services

import (
	"context"
	"errors"
	"log"
	"time"

	"chat_application_api/config"
	"chat_application_api/models"
	"chat_application_api/utilis"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const generatedPasswordLength = 10

func Register(req models.RegisterRequest) (*models.UserResponse, error) {
	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Check if email already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("email already exists")
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Check if phone number already exists
	err = collection.FindOne(ctx, bson.M{"phoneNumber": req.PhoneNumber}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("phone number already exists")
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	plainPassword, err := utilis.GenerateRandomPassword(generatedPasswordLength)
	if err != nil {
		return nil, errors.New("failed to generate password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	now := time.Now()
	user := models.User{
		ID:          primitive.NewObjectID(),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		Age:         req.Age,
		Password:    string(hashedPassword),
		IsOnline:    false,
		LastSeen:    now,
		CreatedAt:   now,
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	if err := SendWelcomeEmail(user.Email, user.FirstName, plainPassword); err != nil {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, delErr := collection.DeleteOne(cleanupCtx, bson.M{"_id": user.ID}); delErr != nil {
			log.Printf("orphaned user after email failure: id=%s cleanupErr=%v", user.ID.Hex(), delErr)
		}
		return nil, errors.New("failed to send welcome email")
	}

	response := &models.UserResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		Age:         user.Age,
		IsOnline:    user.IsOnline,
		LastSeen:    user.LastSeen,
		CreatedAt:   user.CreatedAt,
	}

	return response, nil
}

func RefreshToken(req models.RefreshRequest) (*utilis.TokenPair, error) {
	// Validate the refresh token
	claims, err := utilis.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if refresh token exists in DB
	collection := config.GetCollection("refresh_tokens")
	var storedToken models.RefreshToken
	err = collection.FindOne(ctx, bson.M{"token": req.RefreshToken}).Decode(&storedToken)
	if err != nil {
		return nil, errors.New("refresh token not found or already revoked")
	}

	// Delete the old refresh token (rotate)
	_, err = collection.DeleteOne(ctx, bson.M{"token": req.RefreshToken})
	if err != nil {
		return nil, errors.New("failed to revoke old refresh token")
	}

	// Generate new token pair
	tokens, err := utilis.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Store new refresh token
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	err = storeRefreshToken(ctx, userID, tokens.RefreshToken)
	if err != nil {
		return nil, errors.New("failed to store new refresh token")
	}

	return tokens, nil
}

func storeRefreshToken(ctx context.Context, userID primitive.ObjectID, token string) error {
	collection := config.GetCollection("refresh_tokens")

	refreshToken := models.RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	_, err := collection.InsertOne(ctx, refreshToken)
	return err
}
