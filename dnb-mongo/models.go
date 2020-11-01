package dnb_mongo

import "time"

type TGUser struct {
	ID        string     `bson:"_id"`
	User      *User      `bson:"user"`
	DateTimes *DateTimes `bson:"datetimes"`
	Status    *Status    `bson:"status"`
	Counts    *Counts    `bson:"counts"`
}

type User struct {
	FirstName    string `bson:"first_name"`
	LastName     string `bson:"last_name"`
	Username     string `bson:"username"`
	LanguageCode string `bson:"language_code"`
	Referral     string `bson:"referral"`
	Lang         string `bson:"lang"`
	Bonus        bool   `bson:"bonus"`
}

type DateTimes struct {
	FirstVisit    time.Time `bson:"first_visit"`
	LastVisit     time.Time `bson:"last_visit"`
	Banned        time.Time `bson:"banned"`
	BonusDatetime time.Time `bson:"bonus_datetime"`
}

type Status struct {
	IsAdmin bool `bson:"is_admin"`
	Active  bool `bson:"active"`
}

type Counts struct {
	CountVip      int     `bson:"count_vip"`
	CountFree     int     `bson:"count_free"`
	CountPayments int     `bson:"count_payments"`
	SumSpent      float64 `bson:"sum_spent"`
	Referrals     int     `bson:"referrals"`
	Coins         int     `bson:"coins"`
	Rub           float64 `bson:"rub"`
	BonusCoins    int     `bson:"bonus_coins"`
}
