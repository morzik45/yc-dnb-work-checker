package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	dnb_ydb "yc-dnb-work-checker/dnb-ydb"
	"yc-dnb-work-checker/telegram"
)

func Handler(messages Request) (*Response, error) {
	if len(messages.Messages) == 0 {
		log.Println("Empty messages list")
		return &Response{StatusCode: 200}, nil
	}
	w := new(Work)
	err := json.Unmarshal([]byte(messages.Messages[0].Details.Message.Body), &w)
	if err != nil {
		log.Println("Ошибка при разборе входящих данных:", err)
	}

	yaDB := dnb_ydb.NewDB()
	defer yaDB.Close()

	ws, c, err := yaDB.SetWorkStatus(uint64(w.UID), w.Token)
	if err != nil {
		log.Println("Ошибка при установке статуса:", err.Error())
		return &Response{StatusCode: 200}, nil
	}

	w.WorkType = 0
	w.WorkStatus = int(ws)

	if w.WorkStatus != -1 {
		err = w.AddNewWorkToQueue()
	}
	if err != nil {
		log.Println("Ошибка при добавлении в очередь:", err)
	}

	var text string
	switch w.WorkStatus {
	case -1:
		text = "Исчерпан лимит обработок на сутки!"
	case 0:
		text = fmt.Sprintf("Фото добавлено в очередь.\nУ тебя осталось %d попыток на сегодня.", c)
	case 1:
		text = fmt.Sprintf("Фото добавлено в очередь.\nУ тебя осталось %d монет.", c)
	case 2:
		text = fmt.Sprintf("Фото добавлено в очередь.\nУ тебя осталось %d бонусных монет.", c)
	case 3:
		text = "Фото добавлено в очередь.\nТы админ, какая тебе разница сколько и чего у тебя осталось =)"
	default:
		text = "Что за херня?"
	}
	err = telegram.SendMessage(strconv.Itoa(w.UID), w.Token, text)
	if err != nil {
		log.Println("Ошибка при отправке сообщения пользователю:", err)
	}

	return &Response{StatusCode: 200}, nil
}
