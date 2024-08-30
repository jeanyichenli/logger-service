package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func New(mongoClient *mongo.Client) Models {
	client = mongoClient

	return Models{
		LogEntry: LogEntry{},
	}
}

func (l *LogEntry) Insert(entry LogEntry) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error inserting to logs:", err)
		return err
	}

	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	// create timeout context to avoid implement forever
	// This context will be used to control the overall timeout for the function, ensuring it does not run indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println("Error finding all documents:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []*LogEntry

	for cursor.Next(ctx) {
		var item LogEntry

		if err := cursor.Decode(item); err != nil {
			log.Println("Error decoding log into slice:", err)
			return nil, err
		}

		result = append(result, &item)

	}

	return result, nil
}

func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var entry LogEntry
	err = collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	if err := collection.Drop(ctx); err != nil {
		return err
	}

	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	docId := l.ID

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": docId},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: l.Name},
				{Key: "date", Value: l.Data},
				{Key: "updated_at", Value: time.Now()}},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
