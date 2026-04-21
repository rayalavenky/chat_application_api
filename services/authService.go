package services

import (
	"context"
	"errors"
	"fmt"
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
		Role:        "USER",
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

func Login(req models.LoginRequest) (*models.LoginResponse, error) {
	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	tokens, err := utilis.GenerateTokenPair(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	if err := storeRefreshToken(ctx, user.ID, tokens.RefreshToken); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	now := time.Now()
	if _, err := collection.UpdateOne(ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"isOnline": true, "lastSeen": now}},
	); err != nil {
		log.Printf("failed to update online status for user %s: %v", user.ID.Hex(), err)
	}
	user.IsOnline = true
	user.LastSeen = now

	response := &models.LoginResponse{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		PhoneNumber:  user.PhoneNumber,
		Email:        user.Email,
		Age:          user.Age,
		Role:         user.Role,
		IsOnline:     user.IsOnline,
		LastSeen:     user.LastSeen,
		CreatedAt:    user.CreatedAt,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	return response, nil
}

func Logout(userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id %q: %w", userID, err)
	}

	revokeCtx, revokeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer revokeCancel()

	refreshColl := config.GetCollection("refresh_tokens")
	if _, err := refreshColl.DeleteMany(revokeCtx, bson.M{"userId": objectID}); err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}

	statusCtx, statusCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer statusCancel()

	usersColl := config.GetCollection("users")
	now := time.Now()
	if _, err := usersColl.UpdateOne(statusCtx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"isOnline": false, "lastSeen": now}},
	); err != nil {
		log.Printf("failed to update offline status for user %s: %v", userID, err)
	}

	return nil
}
