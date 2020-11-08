package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strconv"
	dnb_ydb "yc-dnb-work-checker/dnb-ydb"
	"yc-dnb-work-checker/telegram"
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

	yaDB := dnb_ydb.NewDB()
	defer yaDB.Close()

	ws, c, err := yaDB.SetWorkStatus(uint64(w.UID), w.Token)
	if err != nil {
		return err
	}

	w.WorkType = 0
	w.WorkStatus = int(ws)

	err = telegram.SendMessage(strconv.Itoa(w.UID), w.Token, fmt.Sprintf("%d", c))
	if err != nil {
		return err
	}

	return nil
}
