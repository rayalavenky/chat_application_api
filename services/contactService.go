package services

import (
	"chat_application_api/config"
	"chat_application_api/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

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
