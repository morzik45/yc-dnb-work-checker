package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strconv"
	"time"
	dnb_mongo "yc-dnb-work-checker/dnb-mongo"
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
	currentUser, err := db.GetUser(ctx)
	if err != nil {
		return err
	}
	if currentUser.Status.IsAdmin { // Если админ
		w.WorkStatus = 3
	} else if currentUser.User.Bonus &&
		currentUser.DateTimes.BonusDatetime.After(time.Now().UTC()) &&
		currentUser.Counts.BonusCoins > 0 { // VIP за бонусные монеты
		_, err := db.CountInc(ctx, "counts.bonus_coins", -1)
		if err != nil {
			return err
		}
		w.WorkStatus = 2
	} else if currentUser.Counts.Coins > 0 { // VIP
		_, err := db.CountInc(ctx, "counts.coins", -1)
		if err != nil {
			return err
		}
		w.WorkStatus = 1
	} else { // Бесплатная
		w.WorkStatus = 0
	}
	wdb.InsertWork(ctx, w.WorkStatus)
	return nil
}
