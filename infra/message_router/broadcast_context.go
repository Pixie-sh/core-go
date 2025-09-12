package message_router

import (
	"github.com/pixie-sh/core-go/infra/message_wrapper"
)

type BroadcastResult struct {
	MessagesPerRelatedChannelID map[string][]message_wrapper.UntypedMessage //map related channels based on BroadcastChannelID. eg: PartyID
	BroadcastChannelID          string                                      //request BroadcastChannelID within RouterContext. eg: RoomID
}

type BroadcastChannel struct {
	ChannelIdentifier string
	Messages          []message_wrapper.UntypedMessage
	UseFinalizer      bool
}

func (c *BroadcastChannel) AddMessages(messages ...message_wrapper.UntypedMessage) *BroadcastChannel {
	c.Messages = append(c.Messages, messages...)
	return c
}

func (c *BroadcastChannel) WithFinalizer(with bool) *BroadcastChannel {
	c.UseFinalizer = with
	return c
}

type BroadcastContext struct {
	// example:
	// channel: "room_<room uui>"
	// messages: [m1, m2, 3]
	//
	// channel: "party_<party uui>"
	// messages: [m3, m5]
	messagePerChannel map[string]*BroadcastChannel
	BroadcastID       string //usually: connection.ID()
	channelKeys       []string
}

func NewBroadcastContext() *BroadcastContext {
	return &BroadcastContext{
		messagePerChannel: make(map[string]*BroadcastChannel),
		channelKeys:       []string{},
	}
}

func (bc *BroadcastContext) GetChannel(broadcastChannelID string) *BroadcastChannel {
	chn, ok := bc.messagePerChannel[broadcastChannelID]
	if ok {
		return chn
	}

	bc.messagePerChannel[broadcastChannelID] = &BroadcastChannel{
		ChannelIdentifier: broadcastChannelID,
		Messages:          []message_wrapper.UntypedMessage{},
		UseFinalizer:      false,
	}
	bc.channelKeys = append(bc.channelKeys, broadcastChannelID)

	chn, _ = bc.messagePerChannel[broadcastChannelID]
	return chn
}

func (bc *BroadcastContext) GetChannelKeys() []string {
	return bc.channelKeys
}
