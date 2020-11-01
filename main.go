package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"yc-dnb-work-checker/telegram"
)

func Handler(messages Request) (*Response, error) {
	if len(messages.Messages) != 0 {
		w := new(Work)
		err := json.Unmarshal([]byte(messages.Messages[0].Details.Message.Body), &w)
		if err != nil {
			log.Println("Ошибка при разборе входящих данных:", err)
		}
		if w.SetStatus() != nil {
			log.Println("Ошибка при установке статуса:", err)
		} else {
			if w.WorkStatus != -1 {
				if w.AddNewWorkToQueue() != nil {
					log.Println("Ошибка при добавлении в очередь:", err)
				} else {

					if telegram.SendMessage(strconv.Itoa(w.UID), w.Token, fmt.Sprintf("Фото добавлено в очередь")) != nil {
						log.Println("Ошибка при отправке сообщения пользователю:", err)
					}
				}
			} else {
				if telegram.SendMessage(strconv.Itoa(w.UID), w.Token, fmt.Sprintf("Исчерпан лимит на сутки")) != nil {
					log.Println("Ошибка при отправке сообщения пользователю:", err)
				}
			}
		}

	} else {
		log.Println("Empty messages list")
	}
	return &Response{StatusCode: 200}, nil
}
