#!/usr/bin/env python3
# Copyright 2023 The ClusterLink Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import sys
import argparse

proj_dir = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname( os.path.abspath(__file__)))))
sys.path.insert(0,f'{proj_dir}')

from demos.utils.common import runcmd, runcmdb, printHeader
from demos.utils.k8s import getPodName, getPodIp, getPodNameIp
from demos.utils.kind import useKindCluster, getKindIp

srcSvc1  = "productpage"
srcSvc2  = "productpage2"
destSvc  = "reviews"
gw1Name  = "gw1"
allowAllPolicy =f"./mtls/allowAll.json"    
gw3Name        = "peer3"
testOutputFolder = f"{proj_dir}/bin/tests/bookinfo" 


def applyFail(geName, type):
    useKindCluster(geName)
    clPod=getPodName("cl-dataplane")
    print(clPod)
    if type == "fail":
        printHeader(f"Failing {geName} network connection")
        runcmd("kubectl delete service cl-dataplane")
    elif type == "start":
        printHeader(f"Restoring {geName} network connection")
        runcmd("kubectl create service nodeport cl-dataplane --tcp=443:443 --node-port=30443")

############################### MAIN ##########################
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Description of your program')
    parser.add_argument('-t','--type', help='Either fail/start', required=False, default="fail")
    args = vars(parser.parse_args())
    type = args["type"]
    
    print(f'Working directory {proj_dir}')
    os.chdir(proj_dir)
    applyFail(gw3Name, type)
    