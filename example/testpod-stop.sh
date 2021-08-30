#!/bin/bash

source common.sh

echo "[] testpod-stop ($1)"

setNamespace

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
runCommand "kubectl delete -f testpod-$system.yaml --namespace $namespace"

runCommand "sleep 20"

runCommand "kubectl delete -f storageclass-$system.yaml --namespace $namespace"
runCommand "kubectl delete -f secret-$system.yaml --namespace $namespace"
runCommand "helm uninstall --namespace $namespace test-release"

banner "Check Resources"

runCommand "kubectl get all --namespace $namespace"
