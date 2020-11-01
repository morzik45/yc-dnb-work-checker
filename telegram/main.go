package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	logger "yc-dnb-work-checker/dnb-logging"
)

func GetChatMember(userID, token string) (User, error) {
	tr := new(ChatMemberResponse)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChatMember", token)
	fmt.Println(userID, token)
	requestBody, err := json.Marshal(map[string]string{
		"chat_id": userID,
		"user_id": userID,
	})
	if err != nil {
		return User{}, errors.New(fmt.Sprintf("ошибка при кодировании json в getChatMember: %s", err))
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return User{}, errors.New(fmt.Sprintf("ошибка при отправке запроса в TG: %s", err))
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.SaveError(fmt.Sprintf("ошибка при закрытии тела ответа от тг: %s", err))
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return User{}, errors.New(fmt.Sprintf("ошибка при расшифровке ответа от TG: %s", err))
	}
	err = json.Unmarshal(body, &tr)
	if !tr.Ok {
		return User{}, errors.New(fmt.Sprintf("TG вернул ошибку: %s", string(body)))
	}
	return tr.Result.User, nil
}

func SendMessage(userID, token, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	fmt.Println(userID, token, text)
	requestBody, err := json.Marshal(map[string]string{
		"chat_id":    userID,
		"text":       text,
		"parse_mode": "HTML",
	})
	if err != nil {
		return errors.New(fmt.Sprintf("ошибка при кодировании json в SendMessage: %s", err))
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return errors.New(fmt.Sprintf("ошибка при отправке запроса в TG: %s", err))
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.SaveError(fmt.Sprintf("ошибка при закрытии тела ответа от тг: %s", err))
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("ошибка при расшифровке ответа от TG: %s", err))
	}
	tr := new(Response)
	err = json.Unmarshal(body, &tr)
	if !tr.Ok {
		return errors.New(fmt.Sprintf("TG вернул ошибку: %s", string(body)))
	}
	return nil
}
