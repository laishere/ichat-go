package dao

import (
	"encoding/json"
	"fmt"
	"ichat-go/model/entity"
	"time"
)

type CallDao interface {
	CreateCall(c *entity.Call)
	FindCallById(callId uint64) *entity.Call
	GetUserIds(callId uint64) []uint64
	GetCallStatus(callId uint64) int
	UpdateCallStatus(callId uint64, status int) error
	UpdateStartTime(callId uint64)
	UpdateEndReasonAndTime(callId uint64, reason int)
	SetHandled(callId, userId uint64)
	IsHandled(callId, userId uint64) bool
}

func callHandledKey(callId uint64) string {
	return fmt.Sprintf("call:handled:%d", callId)
}

type callDao struct {
	tx Tx
}

func (d callDao) CreateCall(c *entity.Call) {
	assertNoError(d.tx.Create(c))
}

func (d callDao) FindCallById(callId uint64) *entity.Call {
	var c entity.Call
	tx := d.tx.First(&c, "call_id = ?", callId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &c
}

func (d callDao) GetUserIds(callId uint64) []uint64 {
	var c entity.Call
	assertNoError(d.tx.First(&c, "call_id = ?", callId))
	var ids []uint64
	_ = json.Unmarshal([]byte(c.Members), &ids)
	return ids
}

func (d callDao) GetCallStatus(callId uint64) int {
	var c entity.Call
	assertNoError(d.tx.First(&c, "call_id = ?", callId))
	return c.Status
}

func (d callDao) UpdateCallStatus(callId uint64, status int) error {
	return d.tx.Model(&entity.Call{}).Where("call_id = ?", callId).Update("status", status).Error
}

func (d callDao) UpdateStartTime(callId uint64) {
	d.tx.Model(&entity.Call{}).Where("call_id = ?", callId).Updates(map[string]interface{}{
		"start_time": time.Now(),
		"status":     entity.CallStatusActive,
	})
}

func (d callDao) UpdateEndReasonAndTime(callId uint64, reason int) {
	d.tx.Model(&entity.Call{}).Where("call_id = ?", callId).Updates(map[string]interface{}{
		"end_time":   time.Now(),
		"end_reason": reason,
		"status":     entity.CallStatusEnd,
	})
	clearCallCache(callId)
}

func clearCallCache(callId uint64) {
	c := rdb()
	c.Del(c.Context(), callHandledKey(callId))
}

func (d callDao) SetHandled(callId, userId uint64) {
	key := callHandledKey(callId)
	c := rdb()
	c.SAdd(c.Context(), key, userId)
}

func (d callDao) IsHandled(callId, userId uint64) bool {
	key := callHandledKey(callId)
	c := rdb()
	return c.SIsMember(c.Context(), key, userId).Val()
}

func NewCallDao(tx Tx) CallDao {
	return &callDao{tx: tx}
}
