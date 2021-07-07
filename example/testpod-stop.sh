#!/bin/bash

source common.sh

echo "[] testpod-stop ($1)"

helmpath=../helm/csi-charts/

if [ -z ${1+x} ]; then
    echo ""
    echo "Usage: testpod-stop [id]";
    echo "Where:"
    echo "   [id] - specifies a string used to clean up a test pod configuration."
    echo ""
    echo "Example: 'testpod-stop system1'"
    echo ""
    exit
else
    system=$1
fi

banner "Delete Resources"
runCommand "kubectl delete -f testpod-$system.yaml"
runCommand "kubectl delete -f storageclass-$system.yaml"
runCommand "kubectl delete -f secret-$system.yaml"
runCommand "helm uninstall test-release"

banner "Check Resources"

runCommand "kubectl get all"
