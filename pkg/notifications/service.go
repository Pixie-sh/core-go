package notifications

import (
	"context"
	"strconv"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/message_router"
	"github.com/pixie-sh/core-go/infra/message_wrapper"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	notificationmodels "github.com/pixie-sh/core-go/pkg/models/notification_models"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/strings"
)

type PushNotificationService interface {
	SendPushNotificationToToken(token string, title string, message string, additionalData map[string]string) (string, error)
	SendPushNotificationToUserId(userID uint64, title string, message string, additionalData map[string]string) (string, error)
}

type EmailService interface {
	Send(ctx context.Context, to string, subject string, body string) (string, error)
}

type SmsService interface {
	Send(number string, message string, params *notificationmodels.SMSParams) (string, error)
}

type service struct {
	PushNotificationService
	EmailService
	SmsService
}

// Broadcast dumb implementation for now. it just calls firebase service
func (l service) Broadcast(_ string, _ string, wrappers ...message_wrapper.UntypedMessage) message_router.BroadcastResult {
	for _, message := range wrappers {
		log := logger.
			With("payload", message.Payload).
			With("payloadType", message.PayloadType)

		notificationPayload, ok := message.Payload.(notificationmodels.NotificationMessage)
		if !ok {
			log.Warn("unable to cast payload to notification message")
			continue
		}

		sent, err := l.Notify(context.Background(), notificationPayload)
		if err != nil {
			log.With("error", err).Error("error sending push notification via firebase")
			continue
		}

		log.Debug("notifications successfully sent to user; %s", sent)
	}

	return message_router.BroadcastResult{}
}

func (l service) BroadcastCtx(ctx *message_router.BroadcastContext) []message_router.BroadcastResult {
	if types.Nil(ctx) {
		logger.Error("BroadcastCtx called with nil context")
		return nil
	}

	result := make([]message_router.BroadcastResult, len(ctx.GetChannelKeys()))
	for _, key := range ctx.GetChannelKeys() {
		chn := ctx.GetChannel(key)
		result = append(result, l.Broadcast(ctx.BroadcastID, chn.ChannelIdentifier, chn.Messages...))
	}

	return result
}

func (l service) BroadcastFinalizer(_ ...message_router.BroadcastResult) error {
	logger.Debug("notifications.Service broadcast finalizer ignored")
	return nil
}

// Notify uses notifications business layer to synchronously send notification
func (l service) Notify(ctx context.Context, notificationPayload notificationmodels.NotificationMessage) (string, error) {
	switch notificationPayload.GetType() {
	case notificationmodels.PushNotificationType:
		pushPayload, ok := notificationPayload.Payload.(notificationmodels.PushNotificationPayload)
		if !ok {
			return "", errors.New("unable to cast payload to notification payload")
		}

		uintChannelID, err := strconv.ParseUint(pushPayload.UserToken, 10, 64)
		croppedMessage := strings.CropString(
			strings.StripString(
				strings.StripString(pushPayload.Message, "<[^>]*>"),
				"&[a-zA-Z0-9]+;"),
			25,
		)
		if err != nil {
			return l.PushNotificationService.SendPushNotificationToToken(
				pushPayload.UserToken,
				pushPayload.Title,
				croppedMessage,
				pushPayload.AdditionalData,
			)
		}

		return l.PushNotificationService.SendPushNotificationToUserId(
			uintChannelID,
			pushPayload.Title,
			croppedMessage,
			pushPayload.AdditionalData,
		)

	case notificationmodels.SMSNotificationType:
		smsPayload, ok := notificationPayload.Payload.(notificationmodels.SMSNotificationPayload)
		if !ok {
			return "", errors.New("unable to cast payload to notification payload")
		}

		msgSent, err := l.SmsService.Send(smsPayload.PhoneNumber, smsPayload.Message, smsPayload.Params)
		if err != nil {
			return "", err
		}

		return msgSent, nil
	case notificationmodels.EmailNotificationType:
		emailPayload, ok := notificationPayload.Payload.(notificationmodels.EmailNotificationPayload)
		if !ok {
			return "", errors.New("unable to cast payload to notification payload")
		}

		for _, to := range notificationPayload.To {
			id, err := l.EmailService.Send(
				ctx,
				to,
				emailPayload.Subject,
				emailPayload.Body,
			)

			// TODO: Rafa save id for later retry
			if err != nil {
				pixiecontext.GetCtxLogger(ctx).Error("unable to send email with service email ID %s", id)
				return "", err
			}
		}

		return "", nil
	default:
		return "", errors.New("unknown notification type %s", notificationPayload.GetType())
	}

}
