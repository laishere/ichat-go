package notification

import (
	"ichat-go/model/dto"
	"ichat-go/model/entity"
)

func send(userId uint64, n Notification) {
	for _, session := range findSessions(userId) {
		session.Send(n)
	}
}

func SendChatMessage(userId uint64, m *dto.ChatMessageDto, new bool) {
	n := &dto.NotificationMessageDto{
		ChatMessageDto: *m,
		IsNew:          new,
	}
	send(userId, newChatMessage(n))
}

func SendNewContact(userId uint64, c *dto.ContactDto) {
	send(userId, newContact(c))
}

func SendNewContactRequest(userId uint64, r *entity.ContactRequest) {
	send(userId, newContactRequest(r))
}

func SendCallHandled(userId uint64, callId uint64) {
	send(userId, callHandled(callId))
}
