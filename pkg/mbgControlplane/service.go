package mbgControlplane

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.ibm.com/mbg-agent/cmd/mbg/state"
	"github.ibm.com/mbg-agent/pkg/eventManager"
	md "github.ibm.com/mbg-agent/pkg/mbgDataplane"
	"github.ibm.com/mbg-agent/pkg/protocol"
	httpAux "github.ibm.com/mbg-agent/pkg/protocol/http/aux_func"
	"golang.org/x/exp/slices"
)

var slog = logrus.WithField("component", "mbgControlPlane/AddService")

/******************* Local Service ****************************************/
func AddLocalService(s protocol.ServiceRequest) {
	state.UpdateState()
	state.AddLocalService(s.Id, s.Ip, s.Port, s.Description)
}

func GetLocalService(svcId string) protocol.ServiceRequest {
	state.UpdateState()
	s := state.GetLocalService(svcId)
	return protocol.ServiceRequest{Id: s.Id, Ip: s.Ip, Port: s.Port, Description: s.Description}
}

func GetAllLocalServices() map[string]protocol.ServiceRequest {
	state.UpdateState()
	sArr := make(map[string]protocol.ServiceRequest)

	for _, s := range state.GetLocalServicesArr() {
		sPort := state.GetConnectionArr()[s.Id].External
		sIp := state.GetMyIp()
		sArr[s.Id] = protocol.ServiceRequest{Id: s.Id, Ip: sIp, Port: sPort, Description: s.Description}
	}

	return sArr
}

func DelLocalService(svcId string) {
	state.UpdateState()
	svc := state.GetLocalService(svcId)
	mbg := state.GetMyId()
	for _, peer := range svc.PeersExposed {
		peerIp := state.GetMbgTarget(peer)
		delServiceInPeerReq(svcId, mbg, peerIp)
	}
	state.DelLocalService(svcId)
}

func DelLocalServiceFromPeer(svcId, peer string) {
	state.UpdateState()
	svc := state.GetLocalService(svcId)
	mbg := state.GetMyId()
	if slices.Contains(svc.PeersExposed, peer) {
		peerIp := state.GetMbgTarget(peer)
		delServiceInPeerReq(svcId, mbg, peerIp)
	}
	state.DelPeerLocalService(svcId, peer)
}

func delServiceInPeerReq(svcId, serviceMbg, peerIp string) {
	address := state.GetAddrStart() + peerIp + "/remoteservice/" + svcId
	j, err := json.Marshal(protocol.ServiceRequest{Id: svcId, MbgID: serviceMbg})
	if err != nil {
		fmt.Printf("Unable to marshal json: %v", err)
	}

	//send
	resp, _ := httpAux.HttpDelete(address, j, state.GetHttpClient())
	fmt.Printf("Response message for deleting service [%s]:%s \n", svcId, string(resp))
}

/******************* Remote Service ****************************************/

func createRemoteServiceEndpoint(svcId string, force bool) error {
	connPort, err := state.GetFreePorts(svcId)
	if err != nil {
		if err.Error() != state.ConnExist {
			return err
		}
		if !force {
			return nil
		}
	}
	rootCA, certFile, keyFile := state.GetMyMbgCerts()
	slog.Infof("Starting an service endpoint for Remote service %s at port %s with certs(%s,%s,%s)", svcId, connPort.Local, rootCA, certFile, keyFile)
	go md.CreateProxyRemoteService(svcId, connPort.Local, rootCA, certFile, keyFile)
	return nil
}

func RestoreRemoteServices() {
	for svcId, svcArr := range state.GetRemoteServicesArr() {
		allow := false
		for _, svc := range svcArr {
			policyResp, err := state.GetEventManager().RaiseNewRemoteServiceEvent(eventManager.NewRemoteServiceAttr{Service: svc.Id, Mbg: svc.MbgId})
			if err != nil {
				slog.Error("unable to raise connection request event", state.GetMyId())
				continue
			}
			if policyResp.Action == eventManager.Deny {
				continue
			}
			allow = true
		}
		// Create service endpoint only if the service from atleast one MBG is allowed as per policy
		if allow {
			createRemoteServiceEndpoint(svcId, true)
		}
	}
}

func AddRemoteService(e protocol.ExposeRequest) {
	policyResp, err := state.GetEventManager().RaiseNewRemoteServiceEvent(eventManager.NewRemoteServiceAttr{Service: e.Id, Mbg: e.MbgID})
	if err != nil {
		slog.Error("unable to raise connection request event ", state.GetMyId())
		return
	}
	if policyResp.Action == eventManager.Deny {
		slog.Errorf("unable to create service endpoint due to policy")
		return
	}
	err = createRemoteServiceEndpoint(e.Id, false)
	if err != nil {
		return
	}
	state.AddRemoteService(e.Id, e.Ip, e.Description, e.MbgID)
}

func GetRemoteService(svcId string) []protocol.ServiceRequest {
	state.UpdateState()
	return convertRemoteService2RemoteReq(svcId)
}

func GetAllRemoteServices() map[string][]protocol.ServiceRequest {
	state.UpdateState()
	sArr := make(map[string][]protocol.ServiceRequest)

	for svcId, _ := range state.GetRemoteServicesArr() {
		sArr[svcId] = convertRemoteService2RemoteReq(svcId)

	}

	return sArr
}
func convertRemoteService2RemoteReq(svcId string) []protocol.ServiceRequest {
	sArr := []protocol.ServiceRequest{}
	for _, s := range state.GetRemoteService(svcId) {
		sPort := state.GetConnectionArr()[s.Id].Local
		sIp := sPort
		sArr = append(sArr, protocol.ServiceRequest{Id: s.Id, Ip: sIp, Port: sPort, MbgID: s.MbgId, Description: s.Description})
	}
	return sArr
}

func DelRemoteService(svcId, mbgId string) {
	state.UpdateState()
	state.DelRemoteService(svcId, mbgId)
}
