#!/bin/bash

source common.sh

echo "[] testpod-start ($1)"

helmpath=../helm/csi-charts/

if [ -z ${1+x} ]; then
    echo ""
    echo "Usage: testpod-start [id]";
    echo "Where:"
    echo "   [id] - specifies a string used to install and run a particular test pod configuration."
    echo ""
    echo "Example: 'testpod-start system1'"
    echo "   1) helm install test-release $helmpath -f $helmpath/values.yaml"
    echo "   2) kubectl create -f secret-system1.yaml"
    echo "   3) kubectl create -f storageclass-system1.yaml"
    echo "   4) kubectl create -f testpod-system1.yaml"
    echo ""
    exit
else
    system=$1
fi

#
# 1) Run helm install using local charts
#
banner "1) Run helm install using ($helmpath)"
runCommand "helm install test-release $helmpath -f $helmpath/values.yaml"

# Verify that the controller and node pods are running
controllerid=$(kubectl get pods | grep controller | awk '{print $1}')
nodeid=$(kubectl get pods | grep node | awk '{print $1}')

counter=1
success=0
continue=1
maxattempts=5

while [ $continue ]
do
    echo ""
    echo "[$counter] Verify ($controllerid) and ($nodeid) are Running"

    runCommand "kubectl get pods -o wide"
    controllerstatus=$(kubectl get pods $controllerid | grep $controllerid | awk '{print $3}')
    nodestatus=$(kubectl get pods $nodeid | grep $nodeid | awk '{print $3}')

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

if [[ "$success" -eq 0 ]]; then
    exit
fi

#
# 2) Create secrets for the CSI Driver
#
secret=seagate-csi-secrets

banner "2) kubectl create -f secret-$system.yaml"
runCommand "kubectl create -f secret-$system.yaml"
runCommand "kubectl describe secrets $secret"

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

banner "3) kubectl create -f storageclass-$system.yaml"
runCommand "kubectl create -f storageclass-$system.yaml"
runCommand "kubectl describe sc $storageclass"

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

banner "4) kubectl create -f testpod-$system.yaml"
runCommand "kubectl create -f testpod-$system.yaml"

counter=1
success=0
continue=1
maxattempts=15

while [ $continue ]
do
    echo ""
    echo "[$counter] Verify ($testpod) is Running"

    runCommand "kubectl get pods -o wide"
    testpodstatus=$(kubectl get pods $testpod | grep $testpod | awk '{print $3}')

    if [ "$testpodstatus" == "Running" ]; then
        echo "SUCCESS: ($testpod) is Running"
        success=1
        break
    fi

    if [[ "$counter" -eq $maxattempts ]]; then
        echo ""
        echo "ERROR: Max attempts ($maxattempts) reached and ($testpod) is NOT running."
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
