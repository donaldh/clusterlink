C1=left
C2=right
export PROJECT_DIR=`git rev-parse --show-toplevel`
kind create cluster --name ${C1}
kind create cluster --name ${C2}
for f in cl-controlplane cl-dataplane cl-go-dataplane gwctl mlabbe/iperf3
do
    podman image save $f -o img.tar
    kind load image-archive img.tar --name ${C1}
    kind load image-archive img.tar --name ${C2}
    rm img.tar
done
export TEST_DIR=$PROJECT_DIR/test/iperf3
mkdir -p $TEST_DIR
cd $TEST_DIR
$PROJECT_DIR/bin/cl-adm create fabric
$PROJECT_DIR/bin/cl-adm create peer --name ${C1}
$PROJECT_DIR/bin/cl-adm create peer --name ${C2}

kubectl config use-context kind-${C1}
kubectl apply -f ${C1}/k8s.yaml
kubectl apply -f $PROJECT_DIR/demos/utils/manifests/kind/cl-svc.yaml
kubectl create -f $PROJECT_DIR/demos/iperf3/testdata/manifests/iperf3-client/iperf3-client.yaml
export C1_IP=`kubectl get nodes -o "jsonpath={.items[0].status.addresses[0].address}"`

kubectl config use-context kind-${C2}
kubectl apply -f ${C2}/k8s.yaml
kubectl apply -f $PROJECT_DIR/demos/utils/manifests/kind/cl-svc.yaml
kubectl create -f $PROJECT_DIR/demos/iperf3/testdata/manifests/iperf3-server/iperf3.yaml
kubectl create -f $PROJECT_DIR/demos/iperf3/testdata/manifests/iperf3-server/iperf3-cluster-svc.yaml
export C2_IP=`kubectl get nodes -o "jsonpath={.items[0].status.addresses[0].address}"`

sleep 10

$PROJECT_DIR/bin/gwctl init --id "${C1}" \
                       --gwIP $C1_IP --gwPort 30443 \
                       --certca $TEST_DIR/cert.pem \
                       --cert $TEST_DIR/${C1}/gwctl/cert.pem \
                       --key $TEST_DIR/${C1}/gwctl/key.pem
$PROJECT_DIR/bin/gwctl init --id "${C2}" \
                       --gwIP $C2_IP --gwPort 30443 \
                       --certca $TEST_DIR/cert.pem \
                       --cert $TEST_DIR/${C2}/gwctl/cert.pem \
                       --key $TEST_DIR/${C2}/gwctl/key.pem
$PROJECT_DIR/bin/gwctl create peer --myid ${C1} --name ${C2} --host $C2_IP --port 30443
$PROJECT_DIR/bin/gwctl create peer --myid ${C2} --name ${C1} --host $C1_IP --port 30443
$PROJECT_DIR/bin/gwctl create export --myid ${C2} --name iperf3-server --host iperf3-server --port 5000
$PROJECT_DIR/bin/gwctl create import --myid ${C1} --name iperf3-server --host iperf3-server --port 5000
$PROJECT_DIR/bin/gwctl create binding --myid ${C1} --import iperf3-server --peer ${C2}

$PROJECT_DIR/bin/gwctl --myid ${C1} create policy --type access \
      --policyFile $PROJECT_DIR/pkg/policyengine/policytypes/examples/allowAll.json
$PROJECT_DIR/bin/gwctl --myid ${C2} create policy --type access \
      --policyFile $PROJECT_DIR/pkg/policyengine/policytypes/examples/allowAll.json

$PROJECT_DIR/bin/gwctl get all --myid ${C1}
$PROJECT_DIR/bin/gwctl get all --myid ${C2}

kubectl config use-context kind-${C1}
#export IPERF3CLIENT=`kubectl get pods -l app=iperf3-client -o custom-columns=:metadata.name --no-headers`"
#kubectl exec -ti $IPERF3CLIENT -- iperf3 -c iperf3-server --port 5000"
