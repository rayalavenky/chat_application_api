package services

import (
	"chat_application_api/config"
	"chat_application_api/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUsers(search models.UserSearch, loggedInUserID string) ([]models.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate cache key
	cacheKey := fmt.Sprintf(
		"users:%s:%s:%s:%s:%s",
		loggedInUserID,
		search.FirstName,
		search.LastName,
		search.Email,
		search.PhoneNumber,
	)

	// =========================
	// Check Redis First
	// =========================
	cachedData, err := config.RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedUsers []models.UserResponse

		if err := json.Unmarshal([]byte(cachedData), &cachedUsers); err == nil {
			fmt.Println("Users fetched from Redis")
			return cachedUsers, nil
		}
	}

	// =========================
	// MongoDB Collections
	// =========================
	userCollection := config.GetCollection("users")
	requestCollection := config.GetCollection("requests")

	loggedInObjectID, err := primitive.ObjectIDFromHex(loggedInUserID)
	if err != nil {
		return nil, errors.New("invalid logged in user ID")
	}

	filter := bson.M{
		"_id": bson.M{
			"$ne": loggedInObjectID,
		},
	}

	if search.PhoneNumber != "" {
		filter["phoneNumber"] = bson.M{
			"$regex":   search.PhoneNumber,
			"$options": "i",
		}
	}

	if search.FirstName != "" {
		filter["firstName"] = bson.M{
			"$regex":   search.FirstName,
			"$options": "i",
		}
	}

	if search.LastName != "" {
		filter["lastName"] = bson.M{
			"$regex":   search.LastName,
			"$options": "i",
		}
	}

	if search.Email != "" {
		filter["email"] = bson.M{
			"$regex":   search.Email,
			"$options": "i",
		}
	}

	// =========================
	// Fetch Users from MongoDB
	// =========================
	cursor, err := userCollection.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("failed to fetch users")
	}
	defer cursor.Close(ctx)

	requestedUsers, err := getSentRequestIDs(ctx, requestCollection, loggedInObjectID)
	if err != nil {
		return nil, err
	}

	incomingRequests, err := getPendingIncomingRequests(ctx, requestCollection, loggedInObjectID)
	if err != nil {
		return nil, err
	}

	var users []models.UserResponse

	for cursor.Next(ctx) {
		var user models.User

		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}

		incomingID, hasIncoming := incomingRequests[user.ID]

		var incomingIDPtr *primitive.ObjectID
		if hasIncoming {
			incomingIDPtr = &incomingID
		}

		users = append(users, models.UserResponse{
			ID:                user.ID,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			PhoneNumber:       user.PhoneNumber,
			Email:             user.Email,
			Age:               user.Age,
			Bio:               user.Bio,
			IsOnline:          user.IsOnline,
			LastSeen:          user.LastSeen,
			CreatedAt:         user.CreatedAt,
			IsRequestSent:     requestedUsers[user.ID],
			IsRequestReceived: hasIncoming,
			IncomingRequestID: incomingIDPtr,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// =========================
	// Store in Redis
	// =========================
	jsonData, err := json.Marshal(users)
	if err == nil {
		err = config.RDB.Set(
			ctx,
			cacheKey,
			jsonData,
			10*time.Minute,
		).Err()

		if err != nil {
			log.Printf("Redis SET error: %v", err)
		}
	}

	fmt.Println("Users fetched from MongoDB")

	return users, nil
}

func getSentRequestIDs(ctx context.Context, requestCollection *mongo.Collection, fromUserID primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
	cursor, err := requestCollection.Find(ctx, bson.M{"fromUserId": fromUserID})
	if err != nil {
		return nil, errors.New("failed to fetch sent requests")
	}
	defer cursor.Close(ctx)

	sent := map[primitive.ObjectID]bool{}
	for cursor.Next(ctx) {
		var request models.UserRequest
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		sent[request.ToUserID] = true
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return sent, nil
}

func getPendingIncomingRequests(ctx context.Context, requestCollection *mongo.Collection, toUserID primitive.ObjectID) (map[primitive.ObjectID]primitive.ObjectID, error) {
	cursor, err := requestCollection.Find(ctx, bson.M{
		"toUserId": toUserID,
		"status":   "pending",
	})
	if err != nil {
		return nil, errors.New("failed to fetch received requests")
	}
	defer cursor.Close(ctx)

	incoming := map[primitive.ObjectID]primitive.ObjectID{}
	for cursor.Next(ctx) {
		var request models.UserRequest
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		incoming[request.FromUserID] = request.RequestID
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return incoming, nil
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
		Bio:         user.Bio,
		Role:        user.Role,
		IsOnline:    user.IsOnline,
		LastSeen:    user.LastSeen,
		CreatedAt:   user.CreatedAt,
	}

	return response, nil
}

func UpdateProfile(userID string, req models.UpdateProfileRequest) (*models.UserResponse, error) {
	collection := config.GetCollection("users")

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"firstName":   req.FirstName,
			"lastName":    req.LastName,
			"phoneNumber": req.PhoneNumber,
			"age":         req.Age,
			"bio":         req.Bio,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return nil, errors.New("failed to update profile")
	}

	// fetch updated user
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &models.UserResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		Age:         user.Age,
		Bio:         user.Bio,
		IsOnline:    user.IsOnline,
		LastSeen:    user.LastSeen,
		CreatedAt:   user.CreatedAt,
	}, nil
}
