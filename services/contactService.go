package services

import (
	"chat_application_api/config"
	"chat_application_api/models"
	"chat_application_api/utilis"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetContacts(userID string, pagination utilis.Pagination) ([]models.Contact, int64, error) {
	collection := config.GetCollection("contacts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, errors.New("invalid user id")
	}

	filter := bson.M{"userId": objectID}

	totalRecords, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.New("failed to count contacts")
	}

	findOptions := options.Find().
		SetSkip(pagination.Skip()).
		SetLimit(pagination.Limit)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, errors.New("failed to fetch contacts")
	}
	defer cursor.Close(ctx)

	var contacts []models.Contact
	for cursor.Next(ctx) {
		var contact models.Contact
		err := cursor.Decode(&contact)
		if err != nil {
			return nil, 0, errors.New("failed to decode contact")
		}
		contacts = append(contacts, contact)
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return contacts, totalRecords, nil
}

func GetOnlineContacts(userID string, pagination utilis.Pagination) ([]models.Contact, int64, error) {
	collection := config.GetCollection("contacts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, errors.New("invalid user id")
	}

	filter := bson.M{
		"userId":   objectID,
		"isOnline": true,
	}

	totalRecords, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.New("failed to count online contacts")
	}

	findOptions := options.Find().
		SetSkip(pagination.Skip()).
		SetLimit(pagination.Limit)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, errors.New("failed to fetch online contacts")
	}
	defer cursor.Close(ctx)

	var contacts []models.Contact
	for cursor.Next(ctx) {
		var contact models.Contact
		err := cursor.Decode(&contact)
		if err != nil {
			return nil, 0, errors.New("failed to decode contact")
		}
		contacts = append(contacts, contact)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return contacts, totalRecords, nil
}
