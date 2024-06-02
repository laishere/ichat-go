package daemon

import "ichat-go/logic/call"

func Run() {
	go call.MonitorLoop()
}
