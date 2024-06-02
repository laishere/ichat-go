package notification

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"ichat-go/di"
	"ichat-go/logging"
	"ichat-go/sched"
	"strconv"
	"time"
)

const userSessionTTL = time.Second * 60

const (
	SessionStateActive   = 1
	SessionStateInactive = 2
)

type Session interface {
	Send(n Notification)
}

func userSessionsKey(userId uint64) string {
	return fmt.Sprintf("notification:sessions:%d", userId)
}

func sessionKey(userId uint64, sessionId string) string {
	return fmt.Sprintf("noti:%d:%s", userId, sessionId)
}

func newSessionId() string {
	return uuid.NewString()
}

func registerSession(userId uint64, sessionId string) {
	c := di.ENV().RDB()
	c.ZRemRangeByScore(c.Context(), userSessionsKey(userId), "-inf", strconv.FormatInt(time.Now().Unix(), 10))
	c.ZAdd(c.Context(), userSessionsKey(userId), &redis.Z{Score: float64(time.Now().Add(userSessionTTL).Unix()), Member: sessionId})
}

func sessionHeartbeat(userId uint64, sessionId string) {
	c := di.ENV().RDB()
	c.ZAdd(c.Context(), userSessionsKey(userId), &redis.Z{Score: float64(time.Now().Add(userSessionTTL).Unix()), Member: sessionId})
}

func findSessionIds(userId uint64) []string {
	c := di.ENV().RDB()
	keys := c.ZRange(c.Context(), userSessionsKey(userId), 0, -1).Val()
	return keys
}

type mqSession struct {
	mq sched.MQ
}

var sessionLogger = logging.NewLogger("notification")

func (s *mqSession) Send(n Notification) {
	err := s.mq.PushIfStateExits(sched.Message{Payload: n.toJson()})
	if err != nil {
		sessionLogger.Error("Failed to send notification:", err)
	}
}

func findSessions(userId uint64) []Session {
	var sessions []Session
	for _, sessionId := range findSessionIds(userId) {
		mq := sched.NewMQ(sessionKey(userId, sessionId))
		if mq.State() != 0 {
			sessions = append(sessions, &mqSession{mq: mq})
		}
	}
	return sessions
}
