package notifications

var Service service

// Set Deprecated. move way of this approach; use the events to trigger notifications
func Set(pushService PushNotificationService, emailService EmailService, smsService SmsService) {
	s := service{}
	s.EmailService = emailService
	s.PushNotificationService = pushService
	s.SmsService = smsService

	Service = s
}
