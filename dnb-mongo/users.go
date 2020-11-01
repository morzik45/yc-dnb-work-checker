package dnb_mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
	"yc-dnb-work-checker/telegram"
)

type UserDB struct {
	mongo.Collection
}

func NewUserDB(ctx context.Context) (UserDB, error) {
	coll, err := NewMongoDB(ctx, "users")
	return UserDB{coll}, err
}

func (m *UserDB) GetUser(ctx context.Context) (*TGUser, error) {
	userID := ctx.Value("userID").(string)
	var result *TGUser
	filter := bson.D{primitive.E{Key: "_id", Value: userID}}
	update := bson.D{primitive.E{Key: "$currentDate", Value: bson.D{primitive.E{Key: "datetimes.last_visit", Value: true}}}}
	if err := m.FindOneAndUpdate(ctx, filter, update).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			result, err = m.CreateUser(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (m *UserDB) CreateUser(ctx context.Context) (*TGUser, error) {
	userID := ctx.Value("userID").(string)
	token := ctx.Value("token").(string)
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

func (m *UserDB) Update(ctx context.Context, update bson.D) (*TGUser, error) {
	userID := ctx.Value("userID").(string)
	filter := bson.D{primitive.E{Key: "_id", Value: userID}}
	var result *TGUser
	err := m.FindOneAndUpdate(ctx, filter, update).Decode(&result)
	return result, err
}
