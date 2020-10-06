package main

import "log"

type Response struct {
	StatusCode int `json:"statusCode"`
}

type MessageStruct struct {
	Body string `json:"body"`
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

func Handler(messages Request) (*Response, error) {
	log.Println(messages)
	return &Response{StatusCode: 200}, nil
}
