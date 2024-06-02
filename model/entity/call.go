package entity

import "time"

const (
	CallStatusNew    = 1
	CallStatusReady  = 2
	CallStatusActive = 3
	CallStatusEnd    = 4
)

const (
	CallEndReasonNormal         = 1
	CallEndReasonRejected       = 2
	CallEndReasonNoAnswer       = 3
	CallEndReasonBusy           = 4
	CallEndReasonLostConnection = 5
	CallEndReasonError          = 6
	CallEndReasonCancelled      = 7
)

type Call struct {
	CallId    uint64     `json:"callId" gorm:"primaryKey"`
	CallerId  uint64     `json:"callerId"`
	MessageId uint64     `json:"messageId"`
	Members   string     `json:"members"`
	Status    int        `json:"status"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	EndReason int        `json:"endReason"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}
