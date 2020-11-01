package main

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
