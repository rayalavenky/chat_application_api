package services

import (
	"chat_application_api/config"
	"chat_application_api/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUsers() ([]models.UserResponse, error) {
	collection := config.GetCollection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.New("failed to fetch users")
	}
	defer cursor.Close(ctx)

	var users []models.UserResponse

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, models.UserResponse{
			ID:          user.ID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
			Age:         user.Age,
			IsOnline:    user.IsOnline,
			LastSeen:    user.LastSeen,
			CreatedAt:   user.CreatedAt,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserByID(userID string) (*models.UserResponse, error) {
	collection := config.GetCollection("users")

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
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
