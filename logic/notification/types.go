package notification

import (
	"encoding/json"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
)

const (
	typeChatMessage       = 1
	typeNewContact        = 2
	typeNewContactRequest = 3
	typeCallHandled       = 4
)

type Notification struct {
	Type    int `json:"type"`
	Payload any `json:"payload"`
}

func (n *Notification) toJson() []byte {
	b, _ := json.Marshal(n)
	return b
}

func newChatMessage(m *dto.NotificationMessageDto) Notification {
	return Notification{Type: typeChatMessage, Payload: m}
}

func newContact(c *dto.ContactDto) Notification {
	return Notification{Type: typeNewContact, Payload: c}
}

func newContactRequest(r *entity.ContactRequest) Notification {
	return Notification{Type: typeNewContactRequest, Payload: r}
}

func callHandled(callId uint64) Notification {
	return Notification{Type: typeCallHandled, Payload: callId}
}
