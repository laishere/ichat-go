package call

import (
	"ichat-go/logging"
	"ichat-go/sched"
	"strconv"
	"time"
)

const managerTTL = time.Second * 60

var monitorDq sched.DQ

func managerHeartbeat(callId uint64) {
	id := strconv.FormatUint(callId, 10)
	monitorDq.Delete(id)
	_ = monitorDq.Delay(managerTTL, sched.Message{Id: id})
}

func clearManagerHeartbeat(callId uint64) {
	id := strconv.FormatUint(callId, 10)
	monitorDq.Delete(id)
}

var logger logging.Logger

func cleanCall(callId uint64) {
	d := NewManagerDelegate(callId)
	mgr := NewManager(d)
	mgr.CleanAfterDied()
}

func MonitorLoop() {
	logger = logging.NewLogger("call:monitor")
	defer func() {
		if err := recover(); err != nil {
			logger.Error("loop panic: ", err)
		}
	}()
	if monitorDq != nil {
		logger.Error("monitor loop already started")
		return
	}
	monitorDq = sched.NewDQ("call:monitor")
	logger.Debug("enter loop")
	for {
		select {
		case m := <-monitorDq.Channel():
			callId, _ := strconv.ParseUint(m.Id, 10, 64)
			logger.Debug("manager died: ", callId)
			cleanCall(callId)
		}
	}
}
