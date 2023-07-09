// ###############################################################
// Name: Simple iperf3  test
// Desc: create 2 kind clusters :
// 1) MBG and iperf3 client
// 2) MBG and iperf3 server
// ##############################################################
package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.ibm.com/mbg-agent/pkg/controlplane/api"
	mbgAux "github.ibm.com/mbg-agent/tests/utils"
	kindAux "github.ibm.com/mbg-agent/tests/utils/kind"
)

const (
	mbgCaCrt = "./mtls/ca.crt"
	//MBG1 parameters
	mbg1DataPort   = "30001"
	mbg1cPort      = "30443"
	mbg1cPortLocal = "443"
	mbg1crt        = "./mtls/mbg1.crt"
	mbg1key        = "./mtls/mbg1.key"
	mbg1Name       = "mbg1"
	gwctl1Name     = "gwctl1"
	mbg1cni        = "default"
	srcSvc         = "iperf3-client"

	//MBG2 parameters
	mbg2DataPort   = "30001"
	mbg2cPort      = "30443"
	mbg2cPortLocal = "443"
	mbg2crt        = "./mtls/mbg2.crt"
	mbg2key        = "./mtls/mbg2.key"
	mbg2Name       = "mbg2"
	gwctl2Name     = "gwctl2"
	mbg2cni        = "default"
	destSvc        = "iperf3-server"
	destPort       = 5000
	kindDestPort   = "30001"
)

var (
	mtlsFolder string = mbgAux.ProjDir + "/tests/utils/"
	folCl      string = mbgAux.ProjDir + "/tests/iperf3/manifests/iperf3-client"
	folSv      string = mbgAux.ProjDir + "/tests/iperf3/manifests/iperf3-server"
)

func main() {
	// call a Python function
	dataplane := "mtls"
	nologfile := false
	mbgAux.SetLog()
	log.Println("Working directory", mbgAux.ProjDir)
	//exec.chdir(proj_dir)
	//clean
	log.Print("Clean old kinds")
	mbgAux.RunCmd("make clean-kind")

	// build docker environment
	mbgAux.PrintHeader("Build docker image")
	mbgAux.RunCmd("make docker-build")
	kindAux.CreateKindMbg(mbg1Name, dataplane, nologfile)
	kindAux.CreateKindMbg(mbg2Name, dataplane, nologfile)

	// //get parameters
	mbg1Ip, _ := kindAux.GetKindIp(mbg1Name)
	mbg2Ip, _ := kindAux.GetKindIp(mbg2Name)

	//set gwctl
	gwctl1, err := api.CreateGwctl(gwctl1Name, mbg1Ip+":"+mbg1cPort, mtlsFolder+mbgCaCrt, mtlsFolder+mbg1crt, mtlsFolder+mbg1key, dataplane)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	gwctl2, err := api.CreateGwctl(gwctl2Name, mbg2Ip+":"+mbg2cPort, mtlsFolder+mbgCaCrt, mtlsFolder+mbg2crt, mtlsFolder+mbg2key, dataplane)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	//Add Peer
	mbgAux.PrintHeader("Add peers and send hello")
	gwctl1.AddPeer(mbg2Name, mbg2Ip, mbg2cPort)
	gwctl1.SendHello()

	//Set iperf3 client
	mbgAux.PrintHeader("Add iperf3 client")
	kindAux.CreateServiceInKind(mbg1Name, srcSvc, "mlabbe/iperf3", folCl+"/"+srcSvc+".yaml")
	srcSvcPod, _ := mbgAux.GetPodNameIp(srcSvc)
	//gwctl1.AddService(srcSvc, "", "", "iperf3 client") //Allow to use all by default

	//Set iperf3 server
	mbgAux.PrintHeader("Add iperf3 server")
	kindAux.CreateServiceInKind(mbg2Name, destSvc, "mlabbe/iperf3", folSv+"/iperf3.yaml")
	destSvcPod, destSvcIp := mbgAux.GetPodNameIp(destSvc)
	destSvcPort := "5000"
	gwctl2.AddService(destSvc, destSvcIp, destSvcPort, "iperf3 server")
	log.Println(srcSvcPod, destSvcPod)

	//Expose service
	mbgAux.PrintHeader("Start expose")
	kindAux.UseKindCluster(mbg2Name)
	gwctl2.ExposeService(destSvc, "")
	svc, _ := gwctl1.GetRemoteServices()
	log.Println(svc[destSvc])

	//bindK8sSvc()
	mbgAux.PrintHeader("Bind a service")
	kindAux.UseKindCluster(mbg1Name)
	gwctl1.CreateServiceEndpoint(destSvc, destPort, destSvc, "default", "mbg")
	time.Sleep(5 * time.Second)
	//iperf3test
	// mbgLocalPort := strings.Split(svc[destSvc][0].Ip, ":")[1]
	// _, mbglocalIp := mbgAux.GetPodNameIp("mbg")
	// mbgAux.RunCmdNoPipe("kubectl exec -i " + srcSvcPod + " -- iperf3 -c " + mbglocalIp + " -p " + mbgLocalPort)
	mbgAux.RunCmdNoPipe("kubectl exec -i " + srcSvcPod + " -- iperf3 -c " + destSvc + " -p " + "5000")

}

// ############################### MAIN ##########################
// if __name__ == "__main__":
//     parser = argparse.ArgumentParser(description='Description of your program')
//     parser.add_argument('-d','--dataplane', help='choose which dataplane to use mtls/tcp', required=False, default="mtls")
//     parser.add_argument('-c','--cni', help='Which cni to use default(kindnet)/flannel/calico/diff (different cni for each cluster)', required=False, default="default")
