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
runCommand "kubectl delete pods test-pod --grace-period=0 --force"
runCommand "kubectl delete secrets seagate-exos-x-csi-secrets"
runCommand "kubectl delete sc systems-storageclass"
runCommand "kubectl delete pvc systems-pvc"

pvc=$(kubectl get pv | grep pvc | awk '{print $1}')
runCommand "kubectl patch pv $pvc -p '{"metadata": {"finalizers": null}}'"
runCommand "kubectl delete pv $pvc --grace-period=0 --force"

runCommand "kubectl delete pods --all"

banner "Check Resources"

runCommand "kubectl get configmaps"
runCommand "kubectl get serviceaccounts"
runCommand "kubectl get daemonsets"
runCommand "kubectl get deployments"
runCommand "kubectl get pvc"
runCommand "kubectl get pvc"
runCommand "kubectl get sc"
runCommand "kubectl get secrets"
runCommand "kubectl get pods"
