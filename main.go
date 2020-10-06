package main

import (
	"encoding/json"
	"log"
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
	WID    string `json:"wid"`
	UID    int    `json:"uid"`
	Token  string `json:"token"`
	FileID string `json:"file_id"`
}

func Handler(messages Request) (*Response, error) {
	if len(messages.Messages) != 0 {
		w := new(Work)
		json.Unmarshal([]byte(messages.Messages[0].Details.Message.Body), &w)

		log.Println(w.UID)
		log.Println(w.WID)
		log.Println(w.Token)
		log.Println(w.FileID)

	} else {
		log.Println("Empty messages list")
	}
	return &Response{StatusCode: 200}, nil
}
