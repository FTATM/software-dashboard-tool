package model

import "time"

type UserNotification struct {
	UserId      int       `json:"userId" db:"user_id"`
	EmailActive bool      `json:"emailActive" db:"email_active"`
	SmsActive   bool      `json:"smsActive" db:"sms_active"`
	LineActive  bool      `json:"lineActive" db:"line_active"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

type UserNotificationDetail struct {
	UserId        int     `json:"userId" db:"user_id"`
	FirstName     string  `json:"firstName" db:"first_name"`
	LastName      string  `json:"lastName" db:"last_name"`
	Username      string  `json:"username" db:"username"`
	Email         string  `json:"email" db:"email"`
	Tel           string  `json:"tel" db:"tel"`
	LineUserToken *string `json:"lineUserToken" db:"line_user_token"`
	EmailActive   bool    `json:"emailActive" db:"email_active"`
	SmsActive     bool    `json:"smsActive" db:"sms_active"`
	LineActive    bool    `json:"lineActive" db:"line_active"`
}

type UpdateNotification struct {
	UserId      int  `json:"userId"`
	EmailActive bool `json:"emailActive"`
	SmsActive   bool `json:"smsActive"`
	LineActive  bool `json:"lineActive"`
	Active      bool `json:"active"`
}

type UserNotificationSend struct {
	UserId        int     `json:"userId" db:"user_id"`
	EmailActive   bool    `json:"emailActive" db:"email_active"`
	SmsActive     bool    `json:"smsActive" db:"sms_active"`
	LineActive    bool    `json:"lineActive" db:"line_active"`
	Tel           string  `json:"tel" db:"tel"`
	Email         string  `json:"email" db:"email"`
	LineUserToken *string `json:"lineUserToken" db:"line_user_token"`
	Msg           string  `json:"msg"`
}
