package main

import (
	"encoding/json"
	"fmt"
	"log"
	dnb_mongo "yc-dnb-work-checker/dnb-mongo"
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
			if w.AddNewWorkToQueue() != nil {
				log.Println("Ошибка при добавлении в очередь:", err)
			} else {
				db, err := dnb_mongo.NewMongoDB()
				if err != nil {
					log.Println(err)
					return nil, nil
				}
				currentUser, err := db.GetUser(w.WID, w.Token)
				if err != nil {
					fmt.Println(err)
					return nil, nil
				}

				if telegram.SendMessage(w.UID, w.Token, fmt.Sprintf("Фото добавлено в очередь %s, %d", currentUser.User.Username, currentUser.Counts.Coins)) != nil {
					log.Println("Ошибка при отправке сообщения пользователю:", err)
				}
			}
		}

	} else {
		log.Println("Empty messages list")
	}
	return &Response{StatusCode: 200}, nil
}
