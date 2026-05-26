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

func GetUsers(search models.UserSearch) ([]models.UserResponse, error) {
	collection := config.GetCollection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

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

	cursor, err := collection.Find(ctx, filter)
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
			Bio:         user.Bio,
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

func SendRequest(req models.UserRequest) error {
	collection := config.GetCollection("requests")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"fromUserId": req.FromUserID,
		"toUserId":   req.ToUserID,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return errors.New("failed to check existing requests")
	}

	if count > 0 {
		return errors.New("request already exists")
	}

	req.CreatedAt = time.Now()
	req.Status = "pending"

	_, err = collection.InsertOne(ctx, req)
	if err != nil {
		return errors.New("failed to send request")
	}

	return nil
}

func GetSentRequests(userID string) ([]bson.M, error) {
	collection := config.GetCollection("requests")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"fromUserId": objectID,
				},
			},
		},
		{
			{
				Key: "$lookup",
				Value: bson.M{
					"from":         "users",
					"localField":   "toUserId",
					"foreignField": "_id",
					"as":           "toUser",
				},
			},
		},
		{
			{
				Key:   "$unwind",
				Value: "$toUser",
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, errors.New("failed to fetch requests")
	}
	defer cursor.Close(ctx)

	var requests []bson.M

	for cursor.Next(ctx) {
		var request bson.M

		if err := cursor.Decode(&request); err != nil {
			return nil, errors.New("failed to decode request")
		}

		requests = append(requests, request)
	}

	return requests, nil
}

func GetReceivedRequests(userID string) ([]bson.M, error) {
	collection := config.GetCollection("requests")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"toUserId": objectID,
				},
			},
		},
		{
			{
				Key: "$lookup",
				Value: bson.M{
					"from":         "users",
					"localField":   "fromUserId",
					"foreignField": "_id",
					"as":           "fromUser",
				},
			},
		},
		{
			{
				Key:   "$unwind",
				Value: "$fromUser",
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var requests []bson.M

	for cursor.Next(ctx) {
		var request bson.M

		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	return requests, nil
}

func AcceptRequest(requestID string) error {
	requestCollection := config.GetCollection("requests")
	contactCollection := config.GetCollection("contacts")
	userCollection := config.GetCollection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return errors.New("invalid request ID")
	}

	var request models.UserRequest
	err = requestCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("request not found")
		}
		return err
	}

	var fromUser, toUser models.User
	err = userCollection.FindOne(ctx, bson.M{
		"_id": request.FromUserID,
	}).Decode(&fromUser)
	if err != nil {
		return errors.New("sender user not found")
	}

	err = userCollection.FindOne(ctx, bson.M{
		"_id": request.ToUserID,
	}).Decode(&toUser)
	if err != nil {
		return errors.New("receiver user not found")
	}

	// update request status
	_, err = requestCollection.UpdateOne(
		ctx,
		bson.M{
			"_id": objectID,
		},
		bson.M{
			"$set": bson.M{
				"status": "accepted",
			},
		},
	)

	if err != nil {
		return err
	}

	// A -> B
	contact1 := models.Contact{
		UserID:        request.FromUserID,
		ContactUserID: request.ToUserID,
		FirstName:     toUser.FirstName,
		LastName:      toUser.LastName,
		Email:         toUser.Email,
		PhoneNumber:   toUser.PhoneNumber,
		IsOnline:      toUser.IsOnline,
		CreatedAt:     time.Now(),
	}

	// B -> A
	contact2 := models.Contact{
		UserID:        request.ToUserID,
		ContactUserID: request.FromUserID,
		FirstName:     fromUser.FirstName,
		LastName:      fromUser.LastName,
		Email:         fromUser.Email,
		PhoneNumber:   fromUser.PhoneNumber,
		IsOnline:      fromUser.IsOnline,
		CreatedAt:     time.Now(),
	}

	// insert both contacts
	_, err = contactCollection.InsertOne(ctx, contact1)
	if err != nil {
		return err
	}

	_, err = contactCollection.InsertOne(ctx, contact2)
	if err != nil {
		return err
	}

	return nil
}

func GetContacts(userID string) ([]models.Contact, error) {
	collection := config.GetCollection("contacts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"userID": userID})
	if err != nil {
		return nil, errors.New("failed to fetch contacts")
	}
	defer cursor.Close(ctx)

	var contacts []models.Contact
	for cursor.Next(ctx) {
		var contact models.Contact
		err := cursor.Decode(&contact)
		if err != nil {
			return nil, errors.New("failed to decode contact")
		}
		contacts = append(contacts, contact)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return contacts, nil
}
