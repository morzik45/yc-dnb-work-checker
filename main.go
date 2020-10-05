package main

import "log"

type Response struct {
	StatusCode int `json:"statusCode"`
}

type EventMetadataStruct struct {
	EventId   string `json:"event_id"`
	EventType string `json:"event_type"`
	CreatedAt string `json:"created_at"`
}

type MessageStruct struct {
	MessageID string `json:"message_id"`
	MD5OfBody string `json:"md5_of_body"`
	Body      string `json:"body"`
}

type DetailsStruct struct {
	QueueId string        `json:"queue_id"`
	Message MessageStruct `json:"message"`
}

type Message struct {
	EventMetadata EventMetadataStruct `json:"event_metadata"`
	Details       DetailsStruct       `json:"details"`
}

func Handler(messages interface{}) (*Response, error) {
	log.Println(messages)
	return &Response{StatusCode: 200}, nil
}
