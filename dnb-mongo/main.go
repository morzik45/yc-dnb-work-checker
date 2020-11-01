package dnb_mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
	"yc-dnb-work-checker/telegram"
)

type MongoDB struct {
	mongo.Collection
}

func NewMongoDB() (MongoDB, error) {
	var ctx = context.TODO()
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return MongoDB{mongo.Collection{}}, err

	}
	if err := client.Ping(ctx, nil); err != nil {
		return MongoDB{mongo.Collection{}}, err
	}
	Users := client.Database(os.Getenv("DB_NAME")).Collection("users")

	return MongoDB{*Users}, err
}

func (m *MongoDB) GetUser(userID, token string) (*TGUser, error) {
	var ctx = context.TODO()
	var result *TGUser

	filter := bson.D{primitive.E{Key: "_id", Value: userID}}
	update := bson.D{primitive.E{Key: "$currentDate", Value: bson.D{primitive.E{Key: "datetimes.last_visit", Value: true}}}}
	if err := m.FindOneAndUpdate(ctx, filter, update).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			result, err = m.CreateUser(userID, token)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (m *MongoDB) CreateUser(userID, token string) (*TGUser, error) {
	var ctx = context.TODO()

	u, err := telegram.GetChatMember(userID, token)
	if err != nil {
		return nil, err
	}

	var lang string
	if u.Lang == "ru" {
		lang = "ru"
	} else {
		lang = "en"
	}

	newUser := &TGUser{
		ID: userID,
		User: &User{
			FirstName:    u.FirstName,
			LastName:     u.LastName,
			Username:     u.Username,
			LanguageCode: u.Lang,
			Referral:     "",
			Lang:         lang,
			Bonus:        false,
		},
		DateTimes: &DateTimes{
			FirstVisit:    time.Now().UTC(),
			LastVisit:     time.Now().UTC(),
			Banned:        time.Now().UTC().Add(-time.Minute),
			BonusDatetime: time.Now().UTC().Add(-time.Minute),
		},
		Status: &Status{
			IsAdmin: false,
			Active:  true,
		},
		Counts: &Counts{
			CountVip:      0,
			CountFree:     0,
			CountPayments: 0,
			SumSpent:      0,
			Referrals:     0,
			Coins:         0,
			Rub:           0,
			BonusCoins:    0,
		},
	}

	if _, err := m.InsertOne(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}
