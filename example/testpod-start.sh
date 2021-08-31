#!/bin/bash

source common.sh

helmpath=/home/seagate/github.com/Seagate/seagate-exos-x-csi/helm/csi-charts/

# Make sure the helm directory exists
if [ ! -d "$helmpath" ] 
then
    echo ""
    echo "ERROR: Helm path DOES NOT exist, helmpath=$helmpath" 
    echo "NOTE: Update this script with the correct helm path." 
    echo ""
    exit 1
fi

if [ -z ${1+x} ]; then
    echo ""
    echo "Usage: testpod-start [target] [image:version]";
    echo "Where:"
    echo "   [target]        - specifies a string used to install and run a particular test pod configuration."
    echo "   [image:version] - specify a registry image for the csi driver image, overrides values.yaml settings"
    echo ""
    echo "Helm Path: $helmpath"
    echo ""
    echo "Example: testpod-start system1"
    echo "Example: testpod-start system1 docker.io/seagatecsi/seagate-exos-x-csi:v0.5.1"
    echo ""
    echo "Steps:"
    echo "   1) helm install test-release $helmpath -f $helmpath/values.yaml"
    echo "   2) kubectl create -f secret-system1.yaml"
    echo "   3) kubectl create -f storageclass-system1.yaml"
    echo "   4) kubectl create -f testpod-system1.yaml"
    echo ""
    exit
else
    system=$1
fi

if [ -z "$2" ]; then
    registry="default"
else
    registry=$2
    arrIN=(${registry//:/ })
    image=${arrIN[0]}
    version=${arrIN[1]}
fi

setNamespace

echo "[] testpod-start ($system) [$registry] [namespace=$namespace]"

#
# 1) Run helm install using local charts
#
if [ "$registry" == "default" ]; then
    banner "1) Run helm install using ($helmpath)"
    runCommand "helm install test-release --namespace $namespace $helmpath -f $helmpath/values.yaml"
else
    banner "1) Run helm install using ($helmpath) --set image.repository=$image --set image.tag=$version"
    runCommand "helm install test-release --namespace $namespace $helmpath -f $helmpath/values.yaml --set image.repository=$image --set image.tag=$version"
fi

# Verify that the controller and node pods are running
controllerid=$(kubectl get pods --namespace $namespace | grep controller | awk '{print $1}')
nodeid=$(kubectl get pods --namespace $namespace | grep node | awk '{print $1}')

counter=1
success=0
continue=1
maxattempts=6

while [ $continue ]
do
    echo ""
    echo "[$counter] Verify ($controllerid) and ($nodeid) are Running"

    runCommand "kubectl get pods -o wide --namespace $namespace"
    controllerstatus=$(kubectl get pods $controllerid --namespace $namespace | grep $controllerid | awk '{print $3}')
    nodestatus=$(kubectl get pods $nodeid --namespace $namespace | grep $nodeid | awk '{print $3}')

    if [ "$controllerstatus" == "Running" ] && [ "$nodestatus" == "Running" ]; then
        echo "SUCCESS: ($controllerid) and ($nodeid) are Running"
        success=1
        break
    fi

    if [[ "$counter" -eq $maxattempts ]]; then
        echo ""
        echo "ERROR: Max attempts ($maxattempts) reached and pods are not running."
        echo ""
        break
    else
        sleep 5
    fi

    ((counter++))
done

if [ "$success" -eq 0 ]; then
    exit
fi

#
# 2) Create secrets for the CSI Driver
#
secret=seagate-exos-x-csi-secrets

banner "2) kubectl create -f secret-$system.yaml --namespace $namespace"
runCommand "kubectl create -f secret-$system.yaml --namespace $namespace"
runCommand "kubectl describe secrets $secret --namespace $namespace"

if [[ "$?" -ne 0 ]]; then
    echo ""
    echo "ERROR: Secret ($secret) was NOT created successfully."
    echo ""
    exit
fi

#
# 3) Create the Storage Class
#
storageclass=systems-storageclass

banner "3) kubectl create -f storageclass-$system.yaml --namespace $namespace"
runCommand "kubectl create -f storageclass-$system.yaml --namespace $namespace"
runCommand "kubectl describe sc $storageclass --namespace $namespace"

if [[ "$?" -eq 1 ]]; then
    echo ""
    echo "ERROR: StorageClass ($storageclass) was NOT created successfully."
    echo ""
    exit
fi

#
# 4) Create the test pod
#
testpod=test-pod

banner "4) kubectl create -f testpod-$system.yaml --namespace $namespace"
runCommand "kubectl create -f testpod-$system.yaml --namespace $namespace"

counter=1
success=0
continue=1
maxattempts=8

while [ $continue ]
do
    testpodstatus=$(kubectl get pods --namespace $namespace | grep $testpod | awk '{print $3}')
    testpodname=$(kubectl get pods --namespace $namespace | grep $testpod | awk '{print $1}')

    echo ""
    echo "[$counter] Verify ($testpodname) is Running"

    runCommand "kubectl get pods -o wide --namespace $namespace"

    if [ "$testpodstatus" == "Running" ]; then
        echo "SUCCESS: ($testpodname) is Running"
        success=1
        break
    fi

    if [[ "$counter" -eq $maxattempts ]]; then
        echo ""
        echo "ERROR: Max attempts ($maxattempts) reached and ($testpodname) is NOT running."
        echo ""
        break
    else
        sleep 20
    fi

    ((counter++))
done

if [[ "$success" -eq 0 ]]; then
    exit
fi

banner "SUCCESS: All steps succeeded"
