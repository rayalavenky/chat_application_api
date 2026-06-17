package services

import (
	"chat_application_api/config"
	"chat_application_api/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetContacts(userID string) ([]models.Contact, error) {
	collection := config.GetCollection("contacts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	cursor, err := collection.Find(ctx, bson.M{"userId": objectID})
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

func GetOnlineContacts(userID string) ([]models.Contact, error) {
	collection := config.GetCollection("contacts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}
	cursor, err := collection.Find(ctx, bson.M{
		"userId":   objectID,
		"isOnline": true,
	})

	if err != nil {
		return nil, errors.New("failed to fetch online contacts")
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
