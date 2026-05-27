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

	// Validate JWT refresh token
	claims, err := utilis.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Get refresh token from Redis
	redisKey := "refresh:" + claims.UserID

	storedToken, err := config.RDB.Get(
		config.Ctx,
		redisKey,
	).Result()

	if err != nil {
		return nil, errors.New("refresh token not found or expired")
	}

	// Compare stored token with request token
	if storedToken != req.RefreshToken {
		return nil, errors.New("invalid refresh token")
	}

	// Delete old refresh token (rotation)
	err = config.RDB.Del(
		config.Ctx,
		redisKey,
	).Err()

	if err != nil {
		return nil, errors.New("failed to revoke old refresh token")
	}

	// Generate new token pair
	tokens, err := utilis.GenerateTokenPair(
		claims.UserID,
		claims.Email,
	)

	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Convert user id
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// Store new refresh token in Redis
	err = storeRefreshToken(
		userID,
		tokens.RefreshToken,
	)

	if err != nil {
		return nil, errors.New("failed to store new refresh token")
	}

	return tokens, nil
}

func storeRefreshToken(userID primitive.ObjectID, token string) error {

	redisKey := "refresh:" + userID.Hex()

	err := config.RDB.Set(
		config.Ctx,
		redisKey,
		token,
		7*24*time.Hour,
	).Err()

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

	if err := storeRefreshToken(user.ID, tokens.RefreshToken); err != nil {
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

	err := config.RDB.Del(
		config.Ctx,
		"refresh:"+userID,
	).Err()

	if err != nil {
		return err
	}

	return nil
}

func ForgotPassword(req models.ForgotPasswordRequest) error {
	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return errors.New("User not found")
	}

	otp := utilis.GenerateOTP()

	redisKey := "otp:" + req.Email

	err = config.RDB.Set(
		config.Ctx,
		redisKey,
		otp,
		10*time.Minute,
	).Err()

	if err != nil {
		return errors.New("failed to store otp")
	}

	// Send OTP email
	err = SendOTPEmail(user.Email, user.FirstName, otp)
	if err != nil {
		return err
	}

	return nil

}

func VerifyOTP(req models.VerifyOTPRequest) error {

	redisKey := "otp:" + req.Email

	storedOTP, err := config.RDB.Get(
		config.Ctx,
		redisKey,
	).Result()

	if err != nil {
		return errors.New("otp expired or not found")
	}

	if storedOTP != req.OTP {
		return errors.New("invalid otp")
	}

	return nil
}

func ResetPassword(req models.ResetPasswordRequest) error {
	collection := config.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return errors.New("User not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Failed to hash password")
	}

	_, err = collection.UpdateOne(ctx, bson.M{"email": req.Email}, bson.M{
		"$set": bson.M{
			"password":       hashedPassword,
			"resetOTP":       nil,
			"resetOTPExpiry": nil,
		},
	},
	)

	if err != nil {
		return errors.New("Failed to update user")
	}
	config.RDB.Del(config.Ctx, "otp:"+req.Email)

	return nil
}
