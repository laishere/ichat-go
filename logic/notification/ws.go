package notification

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"ichat-go/di"
	"ichat-go/jwt"
	"ichat-go/logging"
	"ichat-go/sched"
	"ichat-go/ws"
	"time"
)

type wsSession struct {
	conn      *websocket.Conn
	mq        sched.MQ
	logger    logging.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	recv      <-chan string
	userId    uint64
	sessionId string
}

func WebSocketHandler(c *gin.Context) {
	conn := ws.Upgrade(c)
	s := newWsSession(conn)
	go s.loop()
}

func newWsSession(conn *websocket.Conn) *wsSession {
	logger := logging.NewLogger("notification_ws")
	ctx, cancel := context.WithCancel(context.Background())
	return &wsSession{conn: conn, logger: logger, ctx: ctx, cancel: cancel}
}

func (s *wsSession) send(n []byte) {
	if err := s.conn.WriteMessage(websocket.TextMessage, n); err != nil {
		panic("Failed to write message: " + err.Error())
	}
}

func (s *wsSession) close() {
	s.logger.Debug("Close session: ", s.sessionId)
	s.cancel()
	_ = s.conn.Close()
	if s.mq != nil {
		// 非常重要，否则会一直消费消息
		s.mq.Close(false)
		s.mq.SaveState(SessionStateInactive)
		s.mq.Expire(userSessionTTL)
	}
}

func (s *wsSession) authenticate() bool {
	select {
	case <-time.After(time.Second * 30):
		return false
	case <-s.ctx.Done():
		return false
	case token := <-s.recv:
		loginId, err := jwt.ValidateToken(token)
		if err != nil {
			return false
		}
		loginUser := di.ENV().LoginUserDao().FindLoginUserByLoginId(loginId)
		if loginUser == nil {
			return false
		}
		s.userId = loginUser.UserId
		return true
	}
}

func (s *wsSession) sessionInit() bool {
	select {
	case <-time.After(time.Second * 30):
		return false
	case <-s.ctx.Done():
		return false
	case sessionId := <-s.recv:
		var mq sched.MQ
		isNewSession := true
		if sessionId != "" {
			mq = sched.NewMQ(sessionKey(s.userId, sessionId))
			mq.ClearExpire()
			if mq.State() == 0 {
				sessionId = ""
			} else {
				isNewSession = false
				sessionHeartbeat(s.userId, sessionId)
				s.logger.Debug("reuse session: ", s.userId, sessionId)
			}
		}
		if sessionId == "" {
			sessionId = newSessionId()
			s.logger.Debug("new session: ", s.userId, sessionId)
			registerSession(s.userId, sessionId)
			mq = sched.NewMQ(sessionKey(s.userId, sessionId))
		}
		reply := "session:" + sessionId
		if isNewSession {
			lastDeliveryId := di.ENV().ChatDao().FindLastDeliveryId(s.userId)
			reply += fmt.Sprintf("\n%d", lastDeliveryId)
		}
		s.send([]byte(reply))
		sessionHeartbeat(s.userId, sessionId)
		mq.SaveState(SessionStateActive)
		mq.Expire(userSessionTTL)
		s.mq = mq
		s.sessionId = sessionId
		s.logger = logging.NewLogger(fmt.Sprintf("notification_ws:%d", s.userId))
		return true
	}
}

func (s *wsSession) loop() {
	defer func() {
		err := recover()
		if err != nil {
			s.logger.Error("loop panic:", err)
		}
		s.close()
	}()
	s.recv = s.read()
	if !s.authenticate() {
		return
	}
	if !s.sessionInit() {
		return
	}
	tick := time.NewTicker(userSessionTTL / 2)
	defer tick.Stop()
	go s.ignoreRead()
	for {
		select {
		case m := <-s.mq.Channel():
			s.send(m.Payload)
			s.mq.Ack(true)
		case <-s.ctx.Done():
			return
		case <-tick.C:
			sessionHeartbeat(s.userId, s.sessionId)
			s.mq.Expire(userSessionTTL)
		}
	}
}

func (s *wsSession) read() <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for {
			_, data, err := s.conn.ReadMessage()
			if err != nil {
				select {
				case <-s.ctx.Done():
					return
				default:
					//s.logger.Debug("Failed to read message: ", err)
					s.cancel()
					return
				}
			}
			ch <- string(data)
		}
	}()
	return ch
}

func (s *wsSession) ignoreRead() {
	for {
		select {
		case <-s.recv:
		case <-s.ctx.Done():
			return
		}
	}
}
