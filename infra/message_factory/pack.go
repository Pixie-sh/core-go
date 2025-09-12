package message_factory

import (
	"github.com/pixie-sh/core-go/infra/message_wrapper"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/utils"
)

type UntypedPackEntry struct {
	MessageType      types.PayloadType             `json:"message_type,omitempty"`
	Descriptions     utils.SchemaDescriptionsModel `json:"descriptions,omitempty"`
	ForceValidations bool
	FromBlob         func(blob []byte) (message_wrapper.UntypedMessage, error) `json:"-"`
	Translate        func(fromPayload any) (any, error)                        `json:"-"`
}

type Pack struct {
	Name    string             `json:"name"`
	Entries []UntypedPackEntry `json:"entries"`
}

func NewPack(name string, entries ...UntypedPackEntry) Pack {
	return Pack{
		name,
		entries,
	}
}

func (p Pack) EntryNames() []string {
	entryNames := make([]string, len(p.Entries))

	for i, entry := range p.Entries {
		entryNames[i] = entry.MessageType.String()
	}

	return entryNames
}
