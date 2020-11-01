package dnb_mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

func NewMongoDB(ctx context.Context, collection string) (mongo.Collection, error) {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return mongo.Collection{}, err

	}
	if err := client.Ping(ctx, nil); err != nil {
		return mongo.Collection{}, err
	}
	coll := client.Database(os.Getenv("DB_NAME")).Collection(collection)

	return *coll, err
}
