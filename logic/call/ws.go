package call

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"ichat-go/di"
	"ichat-go/jwt"
	"ichat-go/logging"
	"ichat-go/model/entity"
	"ichat-go/sched"
	"ichat-go/ws"
	"time"
)

type wsSession struct {
	mq     sched.MQ
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
	logger logging.Logger
	recv   <-chan string
	callId uint64
	userId uint64
}

func WebSocketHandler(c *gin.Context) {
	conn := ws.Upgrade(c)
	logger := logging.NewLogger("call_ws")
	ctx, cancel := context.WithCancel(context.Background())
	s := &wsSession{conn: conn, ctx: ctx, cancel: cancel, logger: logger}
	go s.loop()
}

func GenerateToken(callId uint64, userId uint64) string {
	payload := fmt.Sprintf("call:%d:%d", callId, userId)
	return jwt.GenerateToken(payload, time.Now().Add(time.Hour*3))
}

func validateToken(token string, callId, userId *uint64) bool {
	payload, err := jwt.ValidateToken(token)
	if err != nil {
		return false
	}
	_, err = fmt.Sscanf(payload, "call:%d:%d", callId, userId)
	if err != nil {
		return false
	}
	return true
}

func (s *wsSession) authenticate() bool {
	select {
	case <-s.ctx.Done():
		return false
	case <-time.After(time.Second * 30):
		return false
	case m := <-s.recv:
		var cid, uid uint64
		if !validateToken(m, &cid, &uid) {
			s.send(wsMessage{Type: wsMessageTypeUnauthorized})
			return false
		}
		s.setUserInfo(cid, uid)
		return true
	}
}

func (s *wsSession) setUserInfo(callId uint64, userId uint64) {
	s.callId = callId
	s.userId = userId
	s.logger = logging.NewLogger(fmt.Sprintf("call_ws:%d:%d", callId, userId))
}

func (s *wsSession) send(m wsMessage) {
	err := s.conn.WriteJSON(m)
	if err != nil {
		s.logger.Error("Failed to write message: ", err)
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
					s.logger.Debug("Failed to read message: ", err)
					s.cancel()
					return
				}
			}
			ch <- string(data)
		}
	}()
	return ch
}

func (s *wsSession) close() {
	s.logger.Debug("Close session")
	s.cancel()
	if s.mq != nil {
		s.mq.Close(true)
	}
	if mgr := FindManager(s.callId); mgr != nil {
		// Manager可能已经结束
		mgr.UserOffline(s.userId)
	}
	_ = s.conn.Close()
}

func (s *wsSession) sendError(m string) {
	s.send(wsMessage{Type: wsMessageTypeError, Payload: m})
}

func (s *wsSession) checkCall() bool {
	call := di.ENV().CallDao().FindCallById(s.callId)
	if call == nil {
		s.sendError("通话不存在")
		return false
	}
	if call.Status == entity.CallStatusEnd {
		s.sendError("通话已结束")
		return false
	}
	return true
}

func (s *wsSession) loop() {
	defer func() {
		if err := recover(); err != nil {
			s.logger.Error("Loop panic: ", err)
		}
		s.close()
	}()
	s.recv = s.read()
	if !s.authenticate() {
		return
	}
	if !s.checkCall() {
		return
	}
	mgr := FindManager(s.callId)
	if mgr == nil {
		s.logger.Error("Failed to find manager")
		return
	}
	s.mq = sched.NewMQ(sessionKey(s.callId, s.userId))
	s.mq.SaveState(1)
	mgr.UserOnline(s.userId)
	for {
		select {
		case <-s.ctx.Done():
			return
		case m := <-s.recv:
			var msg wsMessage
			if err := json.Unmarshal([]byte(m), &msg); err != nil {
				s.logger.Error("Failed to unmarshal message: ", err)
				continue
			}
			s.handleMessage(&msg)
		case m := <-s.mq.Channel():
			s.handleAction(m)
			s.mq.Ack(true)
		}
	}
}

func (s *wsSession) handleMessage(m *wsMessage) {
	//s.logger.Debugf("handle message: %d", m.Type)
	mgr := FindManager(s.callId)
	if mgr == nil {
		return
	}
	switch m.Type {
	case wsMessageTypeHeartBeat:
		mgr.HeartBeat(s.userId)
	case wsMessageTypeSignaling:
		sig := signalingPayload(m)
		mgr.Signaling(s.userId, sig.ToUserId, sig.Message)
	default:
		s.logger.Error("Unknown message type: ", m.Type)
	}
}

func (s *wsSession) handleAction(m sched.Message) {
	s.logger.Debugf("action %d", m.Type)
	msg := func(t int) wsMessage {
		return wsMessage{Type: t, Payload: string(m.Payload)}
	}
	switch m.Type {
	case wsActionTypeUpdateUserStates:
		s.send(msg(wsMessageTypeUpdateUserStates))
	case wsActionTypeUpdateUserState:
		s.send(msg(wsMessageTypeUpdateUserState))
	case wsActionTypeSignaling:
		s.send(msg(wsMessageTypeSignaling))
	case wsActionTypeCallStart:
		s.send(wsMessage{Type: wsMessageTypeCallStart})
	case wsActionTypeCallEnd:
		s.send(msg(wsMessageTypeCallEnd))
	case wsActionTypeClose:
		s.cancel()
	default:
		s.logger.Errorf("Unknown action type: %d", m.Type)
	}
}
