package notification_models

type NotificationTypeEnum string

func (e NotificationTypeEnum) String() string {
	return string(e)
}

const (
	SMSNotificationType   NotificationTypeEnum = "SMS"
	PushNotificationType  NotificationTypeEnum = "PUSH"
	EmailNotificationType NotificationTypeEnum = "EMAIL"
)

type NotificationMessage struct {
	NotificationType NotificationTypeEnum `json:"notification_type"`

	To   []string `json:"to"`
	From string   `json:"from"`

	Payload interface{} `json:"payload"`
}

func (n *NotificationMessage) GetType() NotificationTypeEnum {
	return n.NotificationType
}
