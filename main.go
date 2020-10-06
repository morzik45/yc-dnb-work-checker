package main

import "log"

type Response struct {
	StatusCode int `json:"statusCode"`
}

type MessageStruct struct {
	Body Work `json:"body"`
}

type DetailsStruct struct {
	Message MessageStruct `json:"message"`
}

type Message struct {
	Details DetailsStruct `json:"details"`
}

type Request struct {
	Messages []Message `json:"messages"`
}

type Work struct {
	WID    string `json:"wid"`
	UID    int    `json:"uid"`
	Token  string `json:"token"`
	FileID string `json:"file_id"`
}

func Handler(messages Request) (*Response, error) {
	if len(messages.Messages) != 0 {
		log.Println(messages.Messages[0].Details.Message.Body)
		log.Println(messages.Messages[0].Details.Message.Body.UID)
	} else {
		log.Println("Empty messages list")
	}
	return &Response{StatusCode: 200}, nil
}
