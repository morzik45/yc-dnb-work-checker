package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func Handler(messages Request) (*Response, error) {
	if len(messages.Messages) != 0 {
		w := new(Work)
		err := json.Unmarshal([]byte(messages.Messages[0].Details.Message.Body), &w)
		if err != nil {
			log.Println("Ошибка при разборе входящих данных:", err)
		}
		if sendMessage(w.UID, w.Token, fmt.Sprintf("UID: %d\nWID: %s\nToken: %s\nFileID: %s", w.UID, w.WID, w.Token, w.FileID)) != nil {
			log.Println("Ошибка при отправке сообщения пользователю:", err)
		}
	} else {
		log.Println("Empty messages list")
	}
	return &Response{StatusCode: 200}, nil
}
