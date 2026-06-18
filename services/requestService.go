package services

import (
	"chat_application_api/config"
	"chat_application_api/models"
	"chat_application_api/websocket"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SendRequest(req models.UserRequest) error {
	collection := config.GetCollection("requests")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// prevent self request
	if req.FromUserID == req.ToUserID {
		return errors.New("cannot send request to yourself")
	}

	filter := bson.M{
		"fromUserId": req.FromUserID,
		"toUserId":   req.ToUserID,
	}

	var existingRequest models.UserRequest

	err := collection.FindOne(ctx, filter).Decode(&existingRequest)

	if err == nil {

		// already exists
		if existingRequest.Status == "pending" {
			return errors.New("request already sent")
		}

		if existingRequest.Status == "accepted" {
			return errors.New("already connected")
		}

		// rejected -> allow resend
		if existingRequest.Status == "rejected" {

			_, err = collection.UpdateOne(
				ctx,
				filter,
				bson.M{
					"$set": bson.M{
						"status":     "pending",
						"createdAt":  time.Now(),
						"rejectedAt": nil,
						"acceptedAt": nil,
					},
				},
			)

			if err != nil {
				return errors.New("failed to resend request")
			}

			return nil
		}
	}

	// unexpected DB error
	if err != mongo.ErrNoDocuments {
		return errors.New("failed to check existing request")
	}

	// create new request
	req.CreatedAt = time.Now()
	req.Status = "pending"

	_, err = collection.InsertOne(ctx, req)
	if err != nil {
		return errors.New("failed to send request")
	}

	// websocket notification
	websocket.SendToUser(
		req.ToUserID.Hex(),
		map[string]interface{}{
			"type":     "new_request",
			"senderId": req.FromUserID.Hex(),
			"message":  "New connection request",
		},
	)

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
		{
			{
				Key: "$addFields",
				Value: bson.M{
					"requestId": "$_id",
				},
			},
		},
		{
			{
				Key: "$project",
				Value: bson.M{
					"_id": 0,
				},
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
		{
			{
				Key: "$addFields",
				Value: bson.M{
					"requestId": "$_id",
				},
			},
		},
		{
			{
				Key: "$project",
				Value: bson.M{
					"_id": 0,
				},
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

	// update request status (keep the row as history — do not delete)
	_, err = requestCollection.UpdateOne(
		ctx,
		bson.M{
			"_id": objectID,
		},
		bson.M{
			"$set": bson.M{
				"status":     "accepted",
				"acceptedAt": time.Now(),
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

	websocket.SendToUser(
		request.FromUserID.Hex(),
		map[string]interface{}{
			"type":    "request_accepted",
			"userId":  request.ToUserID.Hex(),
			"message": "Your request was accepted",
		},
	)

	return nil
}

func RejectRequest(requestID string) error {
	var request models.UserRequest

	requestCollection := config.GetCollection("requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return errors.New("invalid request ID")
	}

	_, err = requestCollection.UpdateOne(
		ctx,
		bson.M{
			"_id": objectID,
		},
		bson.M{
			"$set": bson.M{
				"status":     "rejected",
				"rejectedAt": time.Now(),
			},
		},
	)
	if err == nil {

		websocket.SendToUser(
			request.FromUserID.Hex(),
			map[string]interface{}{
				"type":    "request_rejected",
				"userId":  request.ToUserID.Hex(),
				"message": "Your request was rejected",
			},
		)
	}
	if err != nil {
		return err
	}
	return nil
}
