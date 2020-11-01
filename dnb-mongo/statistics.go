package dnb_mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type StatisticDB struct {
	mongo.Collection
}

func NewStatisticDB(ctx context.Context) (StatisticDB, error) {
	//collName := time.Now().Format("2006-01-02")
	collName := "works"
	coll, err := NewMongoDB(ctx, collName)
	return StatisticDB{coll}, err
}

func (m *StatisticDB) InsertWork(ctx context.Context, status int) error {
	userID := ctx.Value("userID").(string)

	_, err := m.InsertOne(ctx, bson.D{
		primitive.E{
			Key:   "user_id",
			Value: userID,
		},
		primitive.E{
			Key:   "time",
			Value: time.Now(),
		},
		primitive.E{
			Key:   "status",
			Value: status,
		},
	})

	return err
}
