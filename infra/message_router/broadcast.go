package message_router

import (
	"context"
	"strings"
	"sync"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/events"
	"github.com/pixie-sh/core-go/infra/message_buses"
	"github.com/pixie-sh/core-go/infra/message_wrapper"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/models"
	"github.com/pixie-sh/core-go/pkg/pubsub"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/slices"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type Broadcaster interface {
	Broadcast(broadcastID string, identifier string, message ...message_wrapper.UntypedMessage) BroadcastResult
	BroadcastFinalizer(result ...BroadcastResult) error
	BroadcastCtx(ctx *BroadcastContext) []BroadcastResult
}

type Broadcast struct {
	id                       string
	busPool                  message_buses.BusPool
	appCtx                   context.Context
	subscribeSourceChan      chan SourceSubscription
	informationForConnection func(conn SourceConnection) (SourceInformation, error)
	informationForChannelID  func(channelID string) (SourceInformation, error)
	finalize                 func(performedBroadcasts []BroadcastResult) error
	subscriber               *pubsub.OnProcessSubscriber[SourceSubscription]

	mu      sync.RWMutex
	sources map[string]*struct {
		SourceConnection
		SourceInformation
		SubscriptionID string
	}
}

func (b *Broadcast) ID() string {
	return b.id
}

func NewBroadcast(
	ctx context.Context,
	id string,
	busPool message_buses.BusPool,
	informationForConnection func(conn SourceConnection) (SourceInformation, error),
	informationForChannelID func(channelID string) (SourceInformation, error),
	finalizer func(performedBroadcasts []BroadcastResult) error,
) *Broadcast {
	b := &Broadcast{
		id:                       id,
		busPool:                  busPool,
		appCtx:                   ctx,
		informationForConnection: informationForConnection,
		informationForChannelID:  informationForChannelID,
		finalize:                 finalizer,
		sources: make(map[string]*struct {
			SourceConnection
			SourceInformation
			SubscriptionID string
		}),
	}

	b.subscriber = pubsub.NewOnProcessSubscriber[SourceSubscription](ctx, uid.NewUUID(), b.subscriberHandler, 256)
	return b
}

func (b *Broadcast) SourceSubscriber() pubsub.Subscriber[SourceSubscription] {
	return b.subscriber
}

func (b *Broadcast) subscriberHandler(src SourceSubscription) {
	var (
		connection = src.Connection
		added      = src.Added
	)

	if types.Nil(connection) {
		logger.Logger.Error("subscriber function called with nil connection")
		return
	}

	sourceInformation, err := b.informationForConnection(connection)
	if err != nil {
		logger.Logger.With("err", err).Warn("error fetching connection information: %s", err)
		return
	}

	if len(sourceInformation.ChannelID) == 0 {
		logger.Logger.
			With("connection", connection).
			With("source_information", sourceInformation).
			Error("error fetching connection information, no ChannelID provided")

		return
	}

	if added {
		logger.Logger.Debug("connection %s on broadcaster to be added", connection.ID())
		sourceData := &struct {
			SourceConnection
			SourceInformation
			SubscriptionID string
		}{
			connection,
			sourceInformation,
			"",
		}

		sourceData.SubscriptionID, _ = b.subscribeBuses(
			connection,
			sourceInformation.ChannelID,
			connection.ID(),
		)

		b.mu.Lock()
		b.sources[connection.ID()] = sourceData
		b.mu.Unlock()
		logger.Logger.Debug("connection %s on broadcaster added", connection.ID())
	}

	if !added {
		logger.Logger.Debug("connection %s on broadcaster to be removed", connection.ID())
		b.mu.Lock()
		sourceInfo, ok := b.sources[connection.ID()]
		if !ok {
			return
		}
		delete(b.sources, connection.ID())
		b.mu.Unlock()

		if sourceInfo.SubscriptionID != "" {
			b.unsubscribeBuses(
				sourceInformation.ChannelID,
				sourceInfo.SubscriptionID,
			)
		}
		logger.Logger.Debug("connection %s on broadcaster removed", connection.ID())
	}
}

func (b *Broadcast) BroadcastCtx(ctx *BroadcastContext) []BroadcastResult {
	if types.Nil(ctx) {
		logger.Logger.Error("BroadcastCtx called with nil BroadcastContext")
		return nil
	}

	var (
		broadcastID         = ctx.BroadcastID
		processedBroadcasts []BroadcastResult
		processedBroadcast  BroadcastResult
	)

	for _, channel := range ctx.messagePerChannel {
		if len(channel.Messages) > 0 {
			for _, message := range channel.Messages {
				message.SetHeader(models.HeaderPublisherID, broadcastID)
			}

			processedBroadcast = b.processBroadcast(broadcastID, channel.ChannelIdentifier, channel.Messages, processedBroadcast)
			processedBroadcasts = append(processedBroadcasts, processedBroadcast)

			if !types.IsEmpty(processedBroadcast) && channel.UseFinalizer {
				logger.Logger.Debug("broadcast finalizer for channel %s", channel.ChannelIdentifier)
				err := b.BroadcastFinalizer(processedBroadcast)
				if err != nil {
					logger.Logger.With("error", err).With("broadcast", broadcastID).Error(
						"error finalizing broadcast for %s",
						channel,
					)
				}
				logger.Logger.Debug("broadcast finalizer for channel %s finished", channel.ChannelIdentifier)
			}
		}
	}

	return processedBroadcasts
}

func (b *Broadcast) Broadcast(fromID string, broadcastChannelID string, messages ...message_wrapper.UntypedMessage) BroadcastResult {
	return b.processBroadcast(fromID, broadcastChannelID, messages, BroadcastResult{MessagesPerRelatedChannelID: nil})
}

func (b *Broadcast) processBroadcast(fromID string, broadcastChannelID string, messages []message_wrapper.UntypedMessage, previousResult BroadcastResult) BroadcastResult {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if fromID == broadcastChannelID {
		logger.Logger.Debug("no broadcast needed, destination is the same as origin")
		return BroadcastResult{}
	}

	var alreadyPublishedChannels map[string][]message_wrapper.UntypedMessage
	if previousResult.MessagesPerRelatedChannelID == nil {
		alreadyPublishedChannels = make(map[string][]message_wrapper.UntypedMessage)
	} else {
		alreadyPublishedChannels = previousResult.MessagesPerRelatedChannelID
	}

	sourceInfo, ok := b.sources[broadcastChannelID] //in case broadcastChannelID is a connection.ID
	if ok {
		published := b.subscribeAndPublish(fromID, sourceInfo, broadcastChannelID, messages)
		if published {
			alreadyPublishedChannels[fromID] = messages
		}

		return BroadcastResult{
			MessagesPerRelatedChannelID: alreadyPublishedChannels,
			BroadcastChannelID:          broadcastChannelID,
		}
	}

	alreadyPublishedChannels[fromID] = nil

	//channel ID is a direct broadcast,
	//iterate all over our sources:
	// - channel id doesn't match a connection ID, maybe matches a existing Channel ID
	// - not as primary channel, maybe it's related to channel
	for _, srcInfo := range b.sources {
		srcID := srcInfo.SourceConnection.ID()
		if srcID == fromID {
			logger.Logger.Debug("srcID %s is the same as fromID, ignoring", srcID)
			continue
		}

		_, alreadyPublished := alreadyPublishedChannels[srcInfo.ChannelID]
		if alreadyPublished {
			logger.Logger.Debug("already published %s, ignoring", srcInfo.ChannelID)
			continue
		}

		_, alreadyPublished = alreadyPublishedChannels[srcID]
		if alreadyPublished {
			logger.Logger.Debug("already published %s, ignoring", srcInfo.ChannelID)
			continue
		}

		published := false
		if strings.EqualFold(srcInfo.ChannelID, broadcastChannelID) {
			published = b.subscribeAndPublish(fromID, srcInfo, srcInfo.ChannelID, messages)
		} else if slices.Contains(srcInfo.RelatedTo, broadcastChannelID) {
			published = b.subscribeAndPublish(fromID, srcInfo, srcInfo.ChannelID, messages)
		}

		if published {
			logger.Logger.Debug("for channel %s published to %s", broadcastChannelID, srcInfo.ChannelID)
			alreadyPublishedChannels[srcInfo.ChannelID] = messages
			alreadyPublishedChannels[srcID] = nil
		}
	}

	//the channel ID may not be registered as primary,
	//let's fetch information and iterate over
	broadcastInfo, err := b.informationForChannelID(broadcastChannelID)
	if err != nil {
		logger.Logger.With("error", err).Warn("unable to get extra broadcast information for %s; ignoring broadcast", broadcastChannelID)
		return BroadcastResult{
			MessagesPerRelatedChannelID: alreadyPublishedChannels,
			BroadcastChannelID:          broadcastChannelID,
		}
	}

	logger.Logger.With("broadcastInfo", broadcastInfo).Debug("fetched broadcast info")

	for _, info := range broadcastInfo.RelatedTo {
		_, alreadyPublished := alreadyPublishedChannels[info]
		if alreadyPublished {
			logger.Logger.Debug("already published %s, ignoring", info)
			continue
		}

		// broadcast to existing sources that are related
		for _, broadcastSrcInfo := range b.sources {
			if types.Nil(broadcastSrcInfo.SourceConnection) {
				logger.Logger.Warn("nil connection at broadcast %s == %s", broadcastSrcInfo.ChannelID, info)
				continue
			}

			srcID := broadcastSrcInfo.SourceConnection.ID()
			if srcID == fromID {
				logger.Logger.Debug("srcID %s is the same as fromID, ignoring", srcID)
				continue
			}

			logger.Logger.Debug("trying to broadcast to %s == %s", broadcastSrcInfo.ChannelID, info)
			_, alreadyPublished = alreadyPublishedChannels[srcID]
			if alreadyPublished {
				logger.Logger.Debug("already broadcast %s, ignored", srcID)
				continue
			}

			_, alreadyPublished = alreadyPublishedChannels[broadcastSrcInfo.ChannelID]
			if alreadyPublished {
				logger.Logger.Debug("already broadcast %s, ignored", broadcastSrcInfo.ChannelID)
				continue
			}

			if strings.EqualFold(broadcastSrcInfo.ChannelID, info) {
				logger.Logger.Debug("broadcasting to %s == %s", broadcastSrcInfo.ChannelID, info)
				published := b.subscribeAndPublish(fromID, broadcastSrcInfo, info, messages)
				if published {
					alreadyPublishedChannels[srcID] = nil
					alreadyPublishedChannels[broadcastSrcInfo.ChannelID] = messages
					logger.Logger.Debug("broadcast done to %s == %s", broadcastSrcInfo.ChannelID, info)
				}
			}
			logger.Logger.Debug("finished broadcast to %s == %s", broadcastSrcInfo.ChannelID, info)
		}
	}

	return BroadcastResult{
		MessagesPerRelatedChannelID: alreadyPublishedChannels,
		BroadcastChannelID:          broadcastChannelID,
	}
}

func (b *Broadcast) BroadcastFinalizer(result ...BroadcastResult) error {
	if len(result) == 0 {
		return nil
	}

	return b.finalize(result)
}

// subscribeAndPublish subscribe if not, and publish messages
// NOT LOCKED caller should have lock ownership
// if subscribed the sourceInfo will be affected with SubscriptionID
func (b *Broadcast) subscribeAndPublish(
	fromID string,
	sourceInfo *struct {
		SourceConnection
		SourceInformation
		SubscriptionID string
	},
	channelID string,
	message []message_wrapper.UntypedMessage,
) bool {
	var msgBus message_buses.MessageBus = nil

	if types.Nil(sourceInfo.SourceConnection) {
		logger.Logger.With("srcInfo", sourceInfo).Warn("connection is nil; unable to broadcasting messages")
		return false
	}

	if len(sourceInfo.SubscriptionID) == 0 {
		logger.Logger.Debug("subscribeAndPublish %s == %s", sourceInfo.ChannelID, channelID)
		sourceInfo.SubscriptionID, msgBus = b.subscribeBuses(
			sourceInfo.SourceConnection,
			channelID,
			sourceInfo.SourceConnection.ID())
	} else {
		logger.Logger.Debug("subscribeAndPublish 1 %s == %s", sourceInfo.ChannelID, channelID)
		msgBus = b.busPool.Get(b.appCtx, channelID)
	}

	if types.Nil(msgBus) {
		logger.Logger.With("srcInfo", sourceInfo).With("messages", message).Error("error broadcasting messages, unable to get message bus %s", channelID)
		return false
	}

	logger.Logger.Debug("subscribeAndPublish finished, %s == %s. publishing.", sourceInfo.ChannelID, channelID)
	msgBus.Publish(fromID, message...)
	return true
}

func (b *Broadcast) unsubscribeBuses(channelID string, subscriptionID string) {
	b.busPool.
		Get(b.appCtx, channelID).
		Unsubscribe(subscriptionID)
}

func (b *Broadcast) subscribeBuses(sub pubsub.Subscriber[message_wrapper.UntypedMessage], channelID string, connectionID string) (string, message_buses.MessageBus) {
	bus := b.busPool.
		Get(b.appCtx, channelID)

	return bus.Subscribe(sub), bus
}

// ProduceBatch internally calls Produce for each message
func (b *Broadcast) ProduceBatch(ctx context.Context, wrapper ...events.UntypedEventWrapper) error {
	var errorsList []error

	for _, eventWrapper := range wrapper {
		if types.Nil(eventWrapper) {
			continue
		}

		err := b.Produce(ctx, eventWrapper)
		if err != nil {
			errorsList = append(errorsList, err)
		}
	}

	if len(errorsList) > 0 {
		return errors.New("error producing events").WithErrorCode(errors.ProducerErrorCode).WithNestedError(errorsList...)
	}

	return nil
}

// Produce uses events.UntypedEventWrapper To field to know where to broadcast it
// example: if to[0] == roomID: the message will be propagated to parties on that roomID
// example: if to[1] == partyID: the message will be propagated to that partyID
func (b *Broadcast) Produce(ctx context.Context, wrapper events.UntypedEventWrapper) error {
	destinations := wrapper.To
	wrapper.ClearTo()

	for _, to := range destinations {
		_ = b.Broadcast(wrapper.FromSenderID, to, wrapper.UntypedMessage)
	}

	broadcastTo := wrapper.GetHeaderString("broadcast_to")
	if len(wrapper.To) == 0 && len(broadcastTo) > 0 {
		pixiecontext.GetCtxLogger(ctx).With("message", wrapper).
			Debug("message_router.Broadcast using fallback produce, based on broadcastTo header: %s", broadcastTo)

		_ = b.Broadcast(wrapper.FromSenderID, broadcastTo, wrapper.UntypedMessage)
	}

	return nil
}
