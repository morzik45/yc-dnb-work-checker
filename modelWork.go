package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
	dnb_mongo "yc-dnb-work-checker/dnb-mongo"
	dnb_ydb "yc-dnb-work-checker/dnb-ydb"
)

type Work struct {
	WID        string `json:"wid"`
	UID        int    `json:"uid"`
	Lang       string `json:"lang"`
	Token      string `json:"token"`
	FileID     string `json:"file_id"`
	WorkStatus int    `json:"work_status"`
	WorkType   int    `json:"work_type"`
}

func (w *Work) AddNewWorkToQueue() error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String("ru-central1"),
		Endpoint: aws.String("https://message-queue.api.cloud.yandex.net"),
	}))
	svc := sqs.New(sess)
	queueURL := getEnv("QUEUE_URL", "")
	if queueURL == "" {
		return errors.New("SET QUEUE_URL IN ENV")
	}
	body, err := json.Marshal(w)
	if err != nil {
		return err
	}
	_, err = svc.SendMessage(&sqs.SendMessageInput{
		MessageBody:            aws.String(string(body)),
		QueueUrl:               aws.String(queueURL),
		MessageDeduplicationId: aws.String(w.WID),
	})
	if err != nil {
		return err
	}
	return nil
}

func (w *Work) SetStatus() error {
	w.WorkType = 0    // Работа из TG
	w.WorkStatus = -1 // Изначально отказ
	var ctx = context.Background()
	ctx = context.WithValue(ctx, "userID", strconv.Itoa(w.UID))
	ctx = context.WithValue(ctx, "token", w.Token)
	db, err := dnb_mongo.NewUserDB(ctx)
	if err != nil {
		return err
	}
	wdb, err := dnb_mongo.NewStatisticDB(ctx)
	if err != nil {
		return err
	}

	ydbc := make(chan *dnb_ydb.DB)

	go func(c chan *dnb_ydb.DB) {
		c <- dnb_ydb.NewDB()
	}(ydbc)

	currentUser, err := db.GetUser(ctx)
	if err != nil {
		return err
	}

	if currentUser.Status.IsAdmin { // Если админ
		w.WorkStatus = 3

	} else if currentUser.User.Bonus &&
		currentUser.DateTimes.BonusDatetime.After(time.Now().UTC()) &&
		currentUser.Counts.BonusCoins > 0 { // VIP за бонусные монеты
		update := bson.D{
			primitive.E{
				Key: "$inc",
				Value: bson.D{
					primitive.E{
						Key:   "counts.bonus_coins",
						Value: -1,
					},
					primitive.E{
						Key:   "counts.count_vip",
						Value: 1,
					}},
			}}
		_, err := db.Update(ctx, update)
		if err != nil {
			return err
		}
		w.WorkStatus = 2

	} else if currentUser.Counts.Coins > 0 { // VIP
		update := bson.D{
			primitive.E{
				Key: "$inc",
				Value: bson.D{
					primitive.E{
						Key:   "counts.coins",
						Value: -1,
					},
					primitive.E{
						Key:   "counts.count_vip",
						Value: 1,
					}},
			}}
		_, err := db.Update(ctx, update)
		if err != nil {
			return err
		}
		w.WorkStatus = 1

	} else { // Бесплатная
		t := time.Now().UTC()
		rounded := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		c, err := wdb.CountDocuments(ctx, bson.D{
			primitive.E{
				Key:   "user_id",
				Value: strconv.Itoa(w.UID),
			},
			primitive.E{
				Key:   "status",
				Value: 0,
			},

			primitive.E{
				Key: "time",
				Value: bson.D{
					{
						"$gt",
						rounded,
					}},
			},
		})
		if err != nil {
			return err
		}
		if c < 3 {
			w.WorkStatus = 0
			update := bson.D{
				primitive.E{
					Key: "$inc",
					Value: bson.D{
						primitive.E{
							Key:   "counts.count_free",
							Value: 1,
						}},
				}}
			_, err := db.Update(ctx, update)
			if err != nil {
				return err
			}
		}
	}

	if w.WorkStatus != -1 {
		err = wdb.InsertWork(ctx, w.WorkStatus)
		if err != nil {
			return err
		}
		yaDB := <-ydbc
		defer yaDB.Close()

		err = yaDB.InsertWork(uint64(w.UID), uint8(w.WorkStatus))
		if err != nil {
			return err
		}
	}
	return nil
}
