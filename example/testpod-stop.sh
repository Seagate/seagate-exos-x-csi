#!/bin/bash

source common.sh

echo "[] testpod-stop"

banner "Delete Resources"

runCommand "kubectl delete deployment seagate-exos-x-csi-controller-server"
runCommand "kubectl delete daemonsets seagate-exos-x-csi-node-server"
runCommand "kubectl delete serviceaccounts csi-provisioner"
runCommand "kubectl delete configmaps init-node"
runCommand "kubectl delete clusterrole external-provisioner-runner-systems"
runCommand "kubectl delete ClusterRoleBinding csi-provisioner-role-systems"
runCommand "kubectl delete role external-provisioner-cfg-systems"
runCommand "kubectl delete rolebinding csi-provisioner-role-cfg-systems"

runCommand "helm uninstall test-release"
runCommand "kubectl delete secrets seagate-exos-x-csi-secrets"
runCommand "kubectl delete sc systems-storageclass"

pv=$(kubectl get pv | grep pv | awk '{print $1}')
if [ ! -z "$pv" ]; then
    kubectl patch pv $pv -p '{"metadata": {"finalizers": null}}'
    runCommand "kubectl delete pv $pv --grace-period=0 --force"
fi

pvc=$(kubectl get pv | grep pvc | awk '{print $1}')
if [ ! -z "$pvc" ]; then
    kubectl patch pv $pvc -p '{"metadata": {"finalizers": null}}'
    runCommand "kubectl delete pv $pvc --grace-period=0 --force"
fi

pod=$(kubectl get pod | grep seagate-exos-x-csi-controller | awk '{print $1}')
if [ ! -z "$pod" ]; then
    runCommand "kubectl delete pod $pod --grace-period=0 --force"
fi

pod=$(kubectl get pod | grep seagate-exos-x-csi-node | awk '{print $1}')
if [ ! -z "$pod" ]; then
    runCommand "kubectl delete pod $pod --grace-period=0 --force"
fi

runCommand "kubectl delete pod test-pod --grace-period=0 --force"

banner "Check Resources"

runCommand "kubectl get all"
