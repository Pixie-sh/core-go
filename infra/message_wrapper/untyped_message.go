package message_wrapper

import (
	"time"

	"github.com/pixie-sh/errors-go"
)

type UntypedMessage struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Headers      map[string]interface{} `json:"headers"`
	FromSenderID string                 `json:"from_sender_id"`
	To           []string               `json:"to,omitempty"`
	PayloadType  string                 `json:"payload_type"`

	Payload any      `json:"payload"`
	Error   errors.E `json:"error,omitempty"`
}

func NewFullUntypedMessage(ID string, fromSenderID string, timestamp time.Time, payloadType string, payload any, to ...string) UntypedMessage {
	return UntypedMessage{
		ID:           ID,
		Timestamp:    timestamp,
		Headers:      make(map[string]interface{}),
		FromSenderID: fromSenderID,
		To:           to,
		PayloadType:  payloadType,
		Payload:      payload,
	}
}

func NewUntypedMessage(ID string, payloadType string, payload any) UntypedMessage {
	um := UntypedMessage{
		ID:          ID,
		Headers:     make(map[string]interface{}),
		Timestamp:   time.Now().UTC(),
		PayloadType: payloadType,
		Payload:     payload,
		Error:       nil,
	}
	return um
}

func (w UntypedMessage) Type() string {
	return w.PayloadType
}

func (w UntypedMessage) Data() any {
	return w.Payload
}

func (w *UntypedMessage) AddTo(to string) *UntypedMessage {
	if w.To == nil {
		w.To = []string{to}
	} else {
		w.To = append(w.To, to)
	}

	return w
}

func (w *UntypedMessage) SetError(e error) *UntypedMessage {
	knownError, ok := e.(errors.E)
	if ok {
		w.Error = knownError
	} else {
		w.Error = errors.NewWithError(e, "%s", e.Error())
	}
	return w
}

func (w *UntypedMessage) SetHeader(key string, val interface{}) *UntypedMessage {
	if w.Headers == nil {
		w.Headers = make(map[string]interface{})
	}

	w.Headers[key] = val
	return w
}

func (w *UntypedMessage) GetHeader(key string) interface{} {
	if w.Headers != nil {
		return w.Headers[key]
	}

	return nil
}

func (w *UntypedMessage) GetHeaderString(key string) string {
	if w.Headers != nil {
		str, ok := w.Headers[key].(string)
		if !ok {
			return ""
		}

		return str
	}

	return ""
}

func (w *UntypedMessage) Validate() error {
	if w.ID == "" {
		return errors.New("invalid UntypedMessage: 'ID' field is required")
	}
	//if w.Headers == nil || len(w.Headers) == 0 { // TODO Joao: shall we keep this as nil? maybe all Types should send a data even if {}
	//	return errors.New("invalid UntypedMessage: 'Headers' field is required and can't be empty")
	//}
	if w.FromSenderID == "" {
		return errors.New("invalid UntypedMessage: 'FromSenderID' field is required")
	}
	if w.PayloadType == "" {
		return errors.New("invalid UntypedMessage: 'PayloadType' field is required")
	}
	//if w.Payload == nil { // TODO Joao: shall we keep this as nil? maybe all Types should send a data even if {}
	//	return errors.New("invalid UntypedMessage: 'Payload' field is required")
	//}
	if w.Timestamp.IsZero() {
		return errors.New("invalid UntypedMessage: 'Timestamp' field is required")
	}
	return nil
}

func (w *UntypedMessage) WithFromSenderID(id string) *UntypedMessage {
	w.FromSenderID = id
	return w
}
