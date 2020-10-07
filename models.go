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

type Work struct {
	WID    string `json:"wid"`
	UID    int    `json:"uid"`
	Token  string `json:"token"`
	FileID string `json:"file_id"`
}
