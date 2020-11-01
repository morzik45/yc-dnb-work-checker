package telegram

type Response struct {
	Ok bool `json:"ok"`
}

type ChatMember struct {
	User User `json:"user"`
}

type ChatMemberResponse struct {
	Response
	Result ChatMember `json:"result"`
}

type User struct {
	Id        int    `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Lang      string `json:"language_code"`
}
