package mbgControlplane

import (
	"github.ibm.com/mbg-agent/cmd/mbg/state"
	"github.ibm.com/mbg-agent/pkg/protocol"
)

func Disconnect(d protocol.DisconnectRequest) {
	//Update MBG state
	state.UpdateState()
	connectionID := d.Id + ":" + d.IdDest
	if state.IsServiceLocal(d.IdDest) {
		state.FreeUpPorts(connectionID)
		// Need to Kill the corresponding process
	} else {
		// Need to just Kill the corresponding process
	}

}
