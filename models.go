package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Response struct {
	StatusCode int `json:"statusCode"`
}

type Message struct {
	Body string `json:"body"`
}

type Details struct {
	Message Message `json:"message"`
}

type Item struct {
	Details Details `json:"details"`
}

type Request struct {
	Messages []Item `json:"messages"`
}

type Work struct {
	WID        string `json:"wid"`
	UID        int    `json:"uid"`
	Token      string `json:"token"`
	FileID     string `json:"file_id"`
	WorkStatus int    `json:"work_status"`
	WorkType   int    `json:"work_type"`
}

func (w *Work) SetStatus() error {
	// TODO: Имплементировать выбор и установку статуса и обновление счётчиков
	w.WorkStatus = 2
	w.WorkType = 0
	return nil
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
