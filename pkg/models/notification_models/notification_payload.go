package notification_models

import "time"

type PushNotificationPayload struct {
	UserToken      string            `json:"user_token"`
	Title          string            `json:"title"`
	Message        string            `json:"message"`
	AdditionalData map[string]string `json:"additionalData"`
}

type SMSParams struct {
	GroupIds          []string
	Type              string
	Reference         string
	Validity          int
	Gateway           int
	TypeDetails       map[string]interface{}
	DataCoding        string
	ReportURL         string
	ScheduledDatetime time.Time
	ShortenURLs       bool
}

type SMSNotificationPayload struct {
	PhoneNumber string     `json:"phone_number"`
	Message     string     `json:"message"`
	Params      *SMSParams `json:"parameters"`
}

type EmailNotificationPayload struct {
	Subject string `json:"subject"`
	Body    string `json:"message"`
}
